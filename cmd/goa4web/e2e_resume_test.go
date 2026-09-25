//go:build sqlite

package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/testdata/scenarios"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func extractNonce(body string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return ""
	}
	nonce, _ := doc.Find("input[name='resume_nonce']").Attr("value")
	return nonce
}

func extractResumeToken(location string) string {
	u, err := url.Parse(location)
	if err != nil {
		return ""
	}
	backStr := u.Query().Get("back")
	if backStr == "" {
		return ""
	}
	backURL, err := url.Parse(backStr)
	if err != nil {
		return ""
	}
	return backURL.Query().Get("token")
}

func loginUserFunc(t *testing.T, serverURL, username, password string, client *http.Client) []*http.Cookie {
	req1, _ := http.NewRequest("GET", serverURL+"/login", nil)
	resp1, err := client.Do(req1)
	require.NoError(t, err)
	defer resp1.Body.Close()

	body1, err := io.ReadAll(resp1.Body)
	require.NoError(t, err)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body1)))
	require.NoError(t, err)
	csrf, _ := doc.Find("input[name='gorilla.csrf.Token']").Attr("value")

	form := url.Values{
		"username":           {username},
		"password":           {password},
		"task":               {"Login"},
		"gorilla.csrf.Token": {csrf},
	}

	req2, _ := http.NewRequest("POST", serverURL+"/login", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client.CheckRedirect = nil // enable redirect
	resp2, err := client.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	u, _ := url.Parse(serverURL)
	return client.Jar.Cookies(u)
}

func TestResumeStalePost(t *testing.T) {

	root, err := parseRoot([]string{"serve"})
	if err != nil {
		t.Fatalf("parseRoot: %v", err)
	}
	parent, _ := parseScenarioCmd(root, []string{"serve"})
	serveCmd, _ := parseScenarioServeCmd(parent, []string{"100-private-forum", "-listen", "127.0.0.1:0"})
	serveCmd.fsys = scenarios.FS

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv, _, cleanup, err := serveCmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}
	defer cleanup()

	httpServer := httptest.NewTLSServer(srv.Router)
	defer httpServer.Close()
	serverURL := httpServer.URL
	dbProbe := srv.Queries
	_ = dbProbe

	// We use the browser-like helpers to login

	jarA, _ := cookiejar.New(nil)
	clientA := &http.Client{Jar: jarA, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	core.SessionName = "session"

	// Login UserA (alice)
	loginUserFunc(t, serverURL, "alice", "alice-test", clientA)

	// 2. Render Form
	reqRender, _ := http.NewRequest("GET", serverURL+"/private/topic/new", nil)
	respRender, err := clientA.Do(reqRender)
	require.NoError(t, err)
	bodyRender, err := io.ReadAll(respRender.Body)
	require.NoError(t, err)
	respRender.Body.Close()

	nonce := extractNonce(string(bodyRender))
	require.NotEmpty(t, nonce, "Nonce should be generated on form render")

	// 3. Authentication disappears via real logout
	reqLogout, _ := http.NewRequest("GET", serverURL+"/logout", nil)
	respLogout, _ := clientA.Do(reqLogout)
	respLogout.Body.Close()
	// This will clear the session cookie in clientA's jar, but preserve browser_id

	countBefore, _ := dbProbe.AdminCountForumTopics(context.Background())

	// 4. Submit stale POST
	formStale := url.Values{
		"task":               {"privateTopicCreate"},
		"participants":       {"bob"},
		"title":              {"Secret Plan"},
		"description":        {"Don't tell anyone"},
		"resume_nonce":       {nonce},
		"gorilla.csrf.Token": {"stale-csrf-token"},
	}

	// Disable redirect to catch the 303
	clientA.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	respStale, err := clientA.PostForm(serverURL+"/private/topic/new", formStale)
	require.NoError(t, err)
	defer respStale.Body.Close()
	clientA.CheckRedirect = nil

	assert.Equal(t, http.StatusSeeOther, respStale.StatusCode)
	location := respStale.Header.Get("Location")
	require.Contains(t, location, "/login")
	resumeToken := extractResumeToken(location)
	require.NotEmpty(t, resumeToken, "Resume token should be generated")

	// 5. UserB logs in on a DIFFERENT browser
	jarB, _ := cookiejar.New(nil)
	clientB := &http.Client{Jar: jarB, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	loginUserFunc(t, serverURL, "bob", "bob-test", clientB)

	// UserB attempts to resume UserA's action
	reqIndexGetB, _ := http.NewRequest("GET", serverURL+"/", nil)
	respIndexGetB, _ := clientB.Do(reqIndexGetB)
	bodyIndexGetB, _ := io.ReadAll(respIndexGetB.Body)
	respIndexGetB.Body.Close()
	docIndexGetB, _ := goquery.NewDocumentFromReader(strings.NewReader(string(bodyIndexGetB)))
	csrfResumeB, _ := docIndexGetB.Find("input[name='gorilla.csrf.Token']").Attr("value")

	clientB.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	respBResume, err := clientB.PostForm(serverURL+"/resume", url.Values{"token": {resumeToken}, "gorilla.csrf.Token": {csrfResumeB}})
	require.NoError(t, err)
	respBResume.Body.Close()
	clientB.CheckRedirect = nil
	assert.Equal(t, http.StatusForbidden, respBResume.StatusCode, "UserB should not be able to resume")

	// 6. UserA logs back in on the SAME browser
	loginUserFunc(t, serverURL, "alice", "alice-test", clientA)
	// 7. UserA resumes the action
	reqResumeGet, _ := http.NewRequest("GET", serverURL+"/resume?token="+resumeToken, nil)
	respResumeGet, err := clientA.Do(reqResumeGet)
	require.NoError(t, err)
	bodyResumeGet, err := io.ReadAll(respResumeGet.Body)
	require.NoError(t, err)
	respResumeGet.Body.Close()
	docResumeGet, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyResumeGet)))
	require.NoError(t, err)
	csrfResume, _ := docResumeGet.Find("input[name='gorilla.csrf.Token']").Attr("value")

	clientA.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	respResumeAction, err := clientA.PostForm(serverURL+"/resume", url.Values{"token": {resumeToken}, "gorilla.csrf.Token": {csrfResume}})
	require.NoError(t, err)
	respResumeAction.Body.Close()
	clientA.CheckRedirect = nil

	assert.Equal(t, http.StatusOK, respResumeAction.StatusCode) // TaskDoneAutoRefreshPage

	countAfter, _ := dbProbe.AdminCountForumTopics(context.Background())
	assert.Equal(t, countBefore+1, countAfter, "Exactly one topic should be created after resume")

	// 8. Try to resume AGAIN (should fail)
	clientA.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	respResumeAgain, err := clientA.PostForm(serverURL+"/resume", url.Values{"token": {resumeToken}, "gorilla.csrf.Token": {csrfResume}})
	require.NoError(t, err)
	respResumeAgain.Body.Close()
	assert.Equal(t, http.StatusNotFound, respResumeAgain.StatusCode)
	countAfter2, _ := dbProbe.AdminCountForumTopics(context.Background())
	assert.Equal(t, countAfter, countAfter2, "Topic should not be created twice")

	// --- New Rejection Test for Authorization Revocation ---
	// Create a new client C for user Bob
	jarC, _ := cookiejar.New(nil)
	clientC := &http.Client{Jar: jarC, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	// Bob is a normal user. First, grant Bob access to private forum topics so he can render the form properly.
	_, err = srv.DB.Exec("INSERT INTO grants (user_id, section, item, rule_type, action, item_id) VALUES (2, 'privateforum', 'topic', 'see', 'allow', 0)")
	require.NoError(t, err)

	// Bob logs in
	loginUserFunc(t, serverURL, "bob", "bob-test", clientC)

	// Bob fetches the form and gets a nonce
	reqGetC, _ := http.NewRequest("GET", serverURL+"/private/topic/new", nil)
	respGetC, err := clientC.Do(reqGetC)
	require.NoError(t, err)
	bodyC, _ := io.ReadAll(respGetC.Body)
	respGetC.Body.Close()
	nonceC := extractNonce(string(bodyC))
	require.NotEmpty(t, nonceC, "Bob should be able to get a nonce")

	formC := url.Values{
		"name":         {"Bob Topic"},
		"description":  {"Bob test"},
		"participants": {"alice"},
		"task":         {"privateTopicCreate"},
	}
	formC.Add("resume_nonce", nonceC)

	// Bob explicitly logs out cleanly using the true flow
	reqLogoutGetC, _ := http.NewRequest("GET", serverURL+"/login", nil)
	respLogoutGetC, err := clientC.Do(reqLogoutGetC)
	require.NoError(t, err)
	logoutBodyC, _ := io.ReadAll(respLogoutGetC.Body)
	respLogoutGetC.Body.Close()

	docLogout, _ := goquery.NewDocumentFromReader(strings.NewReader(string(logoutBodyC)))
	logoutCsrfFieldC, _ := docLogout.Find("input[name='gorilla.csrf.Token']").Attr("value")
	require.NotEmpty(t, logoutCsrfFieldC, "Logout CSRF field must exist")

	logoutFormC := url.Values{}
	logoutFormC.Add("gorilla.csrf.Token", logoutCsrfFieldC)
	reqLogoutPostC, _ := http.NewRequest("POST", serverURL+"/usr/logout", strings.NewReader(logoutFormC.Encode()))
	reqLogoutPostC.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respLogoutPostC, err := clientC.Do(reqLogoutPostC)
	require.NoError(t, err)
	respLogoutPostC.Body.Close()

	// Now Bob's session is destroyed. Submit stale POST.
	reqStaleC, _ := http.NewRequest("POST", serverURL+"/private/topic/new", strings.NewReader(formC.Encode()))
	reqStaleC.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	clientC.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	respStaleC, err := clientC.Do(reqStaleC)
	require.NoError(t, err)
	defer respStaleC.Body.Close()
	require.Equal(t, http.StatusSeeOther, respStaleC.StatusCode)
	locC := respStaleC.Header.Get("Location")
	require.Contains(t, locC, "/login")

	resumeTokenC := extractResumeToken(locC)
	require.NotEmpty(t, resumeTokenC)

	// Revoke Bob's authorization before he resumes
	_, err = srv.DB.Exec("DELETE FROM grants WHERE user_id=2 AND section='privateforum' AND item='topic'")
	require.NoError(t, err)

	// Bob logs back in
	clientC.CheckRedirect = nil
	loginUserFunc(t, serverURL, "bob", "bob-test", clientC)

	// Bob fetches the resume page (or just /login) to get a fresh CSRF token
	reqGetUsrC, _ := http.NewRequest("GET", serverURL+"/login", nil)
	respGetUsrC, err := clientC.Do(reqGetUsrC)
	require.NoError(t, err)
	usrBodyC, _ := io.ReadAll(respGetUsrC.Body)
	respGetUsrC.Body.Close()

	docUsr, _ := goquery.NewDocumentFromReader(strings.NewReader(string(usrBodyC)))
	loginCsrfC, _ := docUsr.Find("input[name='gorilla.csrf.Token']").Attr("value")
	require.NotEmpty(t, loginCsrfC)

	// Bob attempts to execute the pending action, which should fail with a 500 error mapped by TaskHandler or 403 because he lost authorization
	// Since we return handlers.ErrForbidden directly from ResumeTaskAction, it propagates through TaskHandler as a standard error page which is HTTP 500 in this framework without specific error mapping overrides.
	reqResumeAuthCheck, _ := http.NewRequest("POST", serverURL+"/resume", strings.NewReader(url.Values{"token": {resumeTokenC}, "gorilla.csrf.Token": {loginCsrfC}}.Encode()))
	reqResumeAuthCheck.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	clientC.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	respResumeAuthCheck, err := clientC.Do(reqResumeAuthCheck)
	require.NoError(t, err)
	respResumeAuthCheck.Body.Close()

	require.Equal(t, http.StatusForbidden, respResumeAuthCheck.StatusCode)

	// Verify the row is still there because it wasn't consumed (we need to hash the token to find it in DB)
	tokenHashC := sha256.Sum256([]byte(resumeTokenC))
	tokenHashHexC := hex.EncodeToString(tokenHashC[:])

	var tokenCount int
	err = srv.DB.QueryRow("SELECT COUNT(*) FROM pending_actions WHERE id = ? AND consumed_at IS NULL", tokenHashHexC).Scan(&tokenCount)
	require.NoError(t, err)
	assert.Equal(t, 1, tokenCount, "Token should remain unconsumed because authorization failed before consumption")

	// Restore the grant
	_, err = srv.DB.Exec("INSERT INTO grants (user_id, section, item, rule_type, action) VALUES (2, 'privateforum', 'topic', 'see', 'see')")
	require.NoError(t, err)
	_, err = srv.DB.Exec("INSERT INTO grants (user_id, section, item, rule_type, action) VALUES (2, 'privateforum', 'topic', 'create', 'create')")
	require.NoError(t, err)

	// Now attempt resumption again, which should succeed
	reqResumeAuthCheck2, _ := http.NewRequest("POST", serverURL+"/resume", strings.NewReader(url.Values{"token": {resumeTokenC}, "gorilla.csrf.Token": {loginCsrfC}}.Encode()))
	reqResumeAuthCheck2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	clientC.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	respResumeAuthCheck2, err := clientC.Do(reqResumeAuthCheck2)
	require.NoError(t, err)
	respResumeAuthCheck2.Body.Close()

	// Note: Since privateTopicCreate uses RefreshDirectHandler, it currently yields a 200 OK meta-refresh
	// instead of a 303 redirect. The PR instructions explicitly say not to convert unrelated routes,
	// so we assert 200 here instead of StatusSeeOther.
	assert.Equal(t, http.StatusOK, respResumeAuthCheck2.StatusCode, "Should succeed after restoring grant (RefreshDirectHandler)")

	// Assert new topic was created (expected because we successfully resumed it)
	// We're adapting the assertion logic down below to expect the original test flow to have NOT created a topic.
	// Since we *did* create one, let's adjust the baseline check here.
	// We'll increment the expectation for the rest of the test...
	countAfter2++

	// Assert new topic was created

	countDAfter, _ := dbProbe.AdminCountForumTopics(context.Background())
	assert.Equal(t, countAfter2, countDAfter, "No new topic should be created on authorization denial")

	// --- New Validation Failure Test (At-Most-Once verification) ---
	// Alice creates a valid POST but with invalid participants (which fails validation AFTER token consumption)
	formInvalid := url.Values{
		"name":         {"Validation Fail Topic"},
		"description":  {"Should burn token without creating"},
		"participants": {"non_existent_user_xyz_123"},
		"task":         {"privateTopicCreate"},
	}

	reqGetInv, _ := http.NewRequest("GET", serverURL+"/private/topic/new", nil)
	respGetInv, err := clientA.Do(reqGetInv)
	require.NoError(t, err)
	bodyInv, _ := io.ReadAll(respGetInv.Body)
	respGetInv.Body.Close()
	nonceInv := extractNonce(string(bodyInv))
	require.NotEmpty(t, nonceInv)

	formInvalid.Add("resume_nonce", nonceInv)

	// Temporarily log out A to trigger intercept
	reqLogoutGetInv, _ := http.NewRequest("GET", serverURL+"/login", nil)
	respLogoutGetInv, _ := clientA.Do(reqLogoutGetInv)
	logoutBodyInv, _ := io.ReadAll(respLogoutGetInv.Body)
	respLogoutGetInv.Body.Close()

	docLogoutInv, _ := goquery.NewDocumentFromReader(strings.NewReader(string(logoutBodyInv)))
	logoutCsrfFieldInv, _ := docLogoutInv.Find("input[name='gorilla.csrf.Token']").Attr("value")

	logoutFormInv := url.Values{}
	logoutFormInv.Add("gorilla.csrf.Token", logoutCsrfFieldInv)
	reqLogoutPostInv, _ := http.NewRequest("POST", serverURL+"/usr/logout", strings.NewReader(logoutFormInv.Encode()))
	reqLogoutPostInv.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respLogoutPostInv, _ := clientA.Do(reqLogoutPostInv)
	respLogoutPostInv.Body.Close()

	// Post the invalid form, which will be intercepted
	clientA.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	reqStaleInv, _ := http.NewRequest("POST", serverURL+"/private/topic/new", strings.NewReader(formInvalid.Encode()))
	reqStaleInv.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respStaleInv, _ := clientA.Do(reqStaleInv)
	respStaleInv.Body.Close()
	locInv := respStaleInv.Header.Get("Location")
	resumeTokenInv := extractResumeToken(locInv)
	require.NotEmpty(t, resumeTokenInv)

	// Log back in as A
	clientA.CheckRedirect = nil
	loginUserFunc(t, serverURL, "alice", "alice-test", clientA)

	reqGetUsrInv, _ := http.NewRequest("GET", serverURL+"/login", nil)
	respGetUsrInv, _ := clientA.Do(reqGetUsrInv)
	usrBodyInv, _ := io.ReadAll(respGetUsrInv.Body)
	respGetUsrInv.Body.Close()

	docUsrInv, _ := goquery.NewDocumentFromReader(strings.NewReader(string(usrBodyInv)))
	loginCsrfInv, _ := docUsrInv.Find("input[name='gorilla.csrf.Token']").Attr("value")

	// Execute the token!
	// It will consume the token, then fail validation, rendering the page instead of redirecting!
	reqResumeInv, _ := http.NewRequest("POST", serverURL+"/resume", strings.NewReader(url.Values{"token": {resumeTokenInv}, "gorilla.csrf.Token": {loginCsrfInv}}.Encode()))
	reqResumeInv.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	clientA.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	respResumeInv, err := clientA.Do(reqResumeInv)
	require.NoError(t, err)
	resumeInvBody, _ := io.ReadAll(respResumeInv.Body)
	respResumeInv.Body.Close()
	assert.Equal(t, http.StatusOK, respResumeInv.StatusCode, "First invalid attempt should return OK with validation errors in body")
	assert.Contains(t, string(resumeInvBody), "Invalid users: non_existent_user_xyz_123")

	// Assert no new topic was created
	countInvAfter, _ := dbProbe.AdminCountForumTopics(context.Background())
	assert.Equal(t, countAfter2, countInvAfter, "No new topic should be created on validation failure")

	// Try to execute the token AGAIN, it should be 404 because it was burned!
	reqResumeInvRetry, _ := http.NewRequest("POST", serverURL+"/resume", strings.NewReader(url.Values{"token": {resumeTokenInv}, "gorilla.csrf.Token": {loginCsrfInv}}.Encode()))
	reqResumeInvRetry.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respResumeInvRetry, _ := clientA.Do(reqResumeInvRetry)
	respResumeInvRetry.Body.Close()

	assert.Equal(t, http.StatusNotFound, respResumeInvRetry.StatusCode, "Token should have been burned by the previous validation failure")
}

// TestResumeNegativePaths tests the negative paths described in the review.
func TestResumeNegativePaths(t *testing.T) {
	root, err := parseRoot([]string{"serve"})
	if err != nil {
		t.Fatalf("parseRoot: %v", err)
	}
	parent, _ := parseScenarioCmd(root, []string{"serve"})
	serveCmd, _ := parseScenarioServeCmd(parent, []string{"100-private-forum", "-listen", "127.0.0.1:0"})
	serveCmd.fsys = scenarios.FS

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv, _, cleanup, err := serveCmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}
	defer cleanup()

	httpServer := httptest.NewTLSServer(srv.Router)
	defer httpServer.Close()
	serverURL := httpServer.URL

	jarA, _ := cookiejar.New(nil)
	clientA := &http.Client{
		Jar: jarA,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// 1. Missing Nonce Validation
	formStaleNoNonce := url.Values{
		"task":               {"privateTopicCreate"},
		"participants":       {"bob"},
		"gorilla.csrf.Token": {"stale-csrf-token"},
	}
	clientA.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	respStale1, err := clientA.PostForm(serverURL+"/private/topic/new", formStaleNoNonce)
	require.NoError(t, err)
	respStale1.Body.Close()
	assert.Equal(t, http.StatusForbidden, respStale1.StatusCode, "Should reject stale POST missing nonce")

	// 2. Unsupported Content-Type Validation
	reqStale2, _ := http.NewRequest("POST", serverURL+"/private/topic/new", strings.NewReader(`{"task": "privateTopicCreate", "resume_nonce": "badnonce"}`))
	reqStale2.Header.Set("Content-Type", "application/json")
	respStale2, err := clientA.Do(reqStale2)
	require.NoError(t, err)
	respStale2.Body.Close()
	assert.Equal(t, http.StatusForbidden, respStale2.StatusCode, "Should reject unsupported content-type")

	// 3. Database persistence error checks. (It's difficult to mock the DB error inside an E2E test without a stub,
	// but the handler returns ErrInternalError which renders a 500 template).
}
