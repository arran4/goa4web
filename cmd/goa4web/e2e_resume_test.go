//go:build sqlite

package main

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/arran4/goa4web/internal/app/server"
	"github.com/arran4/goa4web/testdata/scenarios"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loginUserFunc(t *testing.T, serverURL, username, password string, client *http.Client) {
	reqGet, _ := http.NewRequest("GET", serverURL+"/login", nil)
	respGet, err := client.Do(reqGet)
	require.NoError(t, err)
	body, _ := io.ReadAll(respGet.Body)
	respGet.Body.Close()

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	csrfField, exists := doc.Find("input[name='gorilla.csrf.Token']").Attr("value")
	require.True(t, exists, "CSRF field must exist on login page")

	form := url.Values{}
	form.Add("username", username)
	form.Add("password", password)
	form.Add("gorilla.csrf.Token", csrfField)
	form.Add("task", "Login")

	reqPost, _ := http.NewRequest("POST", serverURL+"/login", strings.NewReader(form.Encode()))
	reqPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respPost, err := client.Do(reqPost)
	require.NoError(t, err)
	respPost.Body.Close()
	if respPost.StatusCode != http.StatusOK && respPost.StatusCode != http.StatusNotFound {
		t.Fatalf("Login failed with status %d", respPost.StatusCode)
	}
}

func extractNonce(body string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return ""
	}
	nonce, _ := doc.Find("input[name='resume_nonce']").Attr("value")
	return nonce
}

func extractResumeToken(loc string) string {
	u, err := url.Parse(loc)
	if err != nil {
		return ""
	}
	backURL, err := url.Parse(u.Query().Get("back"))
	if err != nil {
	    return ""
	}
	return backURL.Query().Get("token")
}

func setupTestServer(t *testing.T) (*httptest.Server, *server.Server, func()) {
	root, err := parseRoot([]string{"serve"})
	if err != nil {
		t.Fatalf("parseRoot: %v", err)
	}
	parent, _ := parseScenarioCmd(root, []string{"serve"})
	serveCmd, _ := parseScenarioServeCmd(parent, []string{"100-private-forum", "-listen", "127.0.0.1:0"})
	serveCmd.fsys = scenarios.FS

	ctx, cancel := context.WithCancel(context.Background())

	srv, _, cleanupServe, err := serveCmd.Bootstrap(ctx)
	if err != nil {
		cancel()
		t.Fatalf("Bootstrap failed: %v", err)
	}

	httpServer := httptest.NewTLSServer(srv.Router)

	cleanup := func() {
		httpServer.Close()
		cleanupServe()
		cancel()
	}

	return httpServer, srv, cleanup
}

func createClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

// 1. Authenticated invalid-CSRF rejection
func TestResume_AuthenticatedInvalidCSRFRejection(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	_ = srv
	serverURL := httpServer.URL

	client := createClient()
	loginUserFunc(t, serverURL, "bob", "bob-test", client)

	// Fetch form
	reqGet, _ := http.NewRequest("GET", serverURL+"/private/topic/new", nil)
	respGet, _ := client.Do(reqGet)
	bodyGet, _ := io.ReadAll(respGet.Body)
	respGet.Body.Close()

	nonce := extractNonce(string(bodyGet))
	if nonce == "" {
		t.Logf("Failed to get nonce. Body: %s", string(bodyGet))
	}
	require.NotEmpty(t, nonce)

		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(bodyGet)))
	formAction, exists := doc.Find("form#private-form").Attr("action")
	require.True(t, exists, "Form action attribute must exist")
	require.Equal(t, "/private/topic/new", formAction, "Form action must be exactly /private/topic/new")
	parsedAction, _ := url.Parse(formAction)
	actionURL := reqGet.URL.ResolveReference(parsedAction)

	formStale := url.Values{
		"task":               {"Private topic create"},
		"participants":       {"bob"},
		"title":              {"Test Topic"},
		"description":        {"Test"},
		"gorilla.csrf.Token": {"stale-csrf-token"}, // Invalid CSRF
		"resume_nonce":       {nonce},
	}

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	reqStale, _ := http.NewRequest("POST", actionURL.String(), strings.NewReader(formStale.Encode()))
	reqStale.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respStale, err := client.Do(reqStale)
	require.NoError(t, err)
	defer respStale.Body.Close()

	assert.Equal(t, http.StatusForbidden, respStale.StatusCode, "Authenticated users with invalid CSRF must be rejected with 403, not captured")

	// Verify no token was created
	var count int
	_ = srv.DB.QueryRow("SELECT COUNT(*) FROM pending_actions").Scan(&count)
	assert.Equal(t, 1, count, "No new pending action should be stored for authenticated user, only the one from GET")
}

// 2. Genuine logout -> stale capture -> same-user resume
func TestResume_GenuineLogoutAndResume(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	_ = srv
	serverURL := httpServer.URL
	dbProbe := srv.Queries
	_ = srv

	client := createClient()

	loginUserFunc(t, serverURL, "alice", "alice-test", client)

	countBefore, _ := dbProbe.AdminCountForumTopics(context.Background())

	reqGet, _ := http.NewRequest("GET", serverURL+"/private/topic/new", nil)
	respGet, _ := client.Do(reqGet)
	bodyGet, _ := io.ReadAll(respGet.Body)
	respGet.Body.Close()
	nonce := extractNonce(string(bodyGet))

		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(bodyGet)))
	formAction, exists := doc.Find("form#private-form").Attr("action")
	require.True(t, exists, "Form action attribute must exist")
	require.Equal(t, "/private/topic/new", formAction, "Form action must be exactly /private/topic/new")
	parsedAction, _ := url.Parse(formAction)
	actionURL := reqGet.URL.ResolveReference(parsedAction)

	// True logout flow
	reqLogoutGet, _ := http.NewRequest("GET", serverURL+"/usr/logout", nil)
	respLogoutGet, _ := client.Do(reqLogoutGet)
	bodyLogoutGet, _ := io.ReadAll(respLogoutGet.Body)
	respLogoutGet.Body.Close()

	require.Equal(t, 200, respLogoutGet.StatusCode)


	docLogout, _ := goquery.NewDocumentFromReader(strings.NewReader(string(bodyLogoutGet)))
	logoutCsrfField, _ := docLogout.Find("form[action='/usr/logout'] input[name='gorilla.csrf.Token']").Attr("value")
			require.NotEmpty(t, logoutCsrfField)

	logoutForm := url.Values{}
	logoutForm.Add("gorilla.csrf.Token", logoutCsrfField)
	reqLogoutPost, _ := http.NewRequest("POST", serverURL+"/usr/logout", strings.NewReader(logoutForm.Encode()))
	reqLogoutPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqLogoutPost.Header.Set("X-CSRF-Token", logoutCsrfField)
	reqLogoutPost.Header.Set("Referer", serverURL+"/usr/logout")
	reqLogoutPost.Header.Set("X-CSRF-Token", logoutCsrfField)

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	respLogoutPost, _ := client.Do(reqLogoutPost)
	require.Equal(t, http.StatusSeeOther, respLogoutPost.StatusCode, "True logout should return 303")
	respLogoutPost.Body.Close()

	// Verify unauthenticated by fetching a protected page
	reqTestAuth, _ := http.NewRequest("GET", serverURL+"/private/topic/new", nil)
	respTestAuth, _ := client.Do(reqTestAuth)
				require.Equal(t, http.StatusForbidden, respTestAuth.StatusCode, "Should be redirected to login because unauthenticated")
	respTestAuth.Body.Close()

	// Submit stale POST
	formStale := url.Values{
		"task":               {"Private topic create"},
		"participants":       {"bob"},
		"title":              {"Bob Topic"},
		"description":        {"Bob test"},
		"gorilla.csrf.Token": {"stale-csrf-token"},
		"resume_nonce":       {nonce},
	}
	reqStale, _ := http.NewRequest("POST", actionURL.String(), strings.NewReader(formStale.Encode()))
	reqStale.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respStale, err := client.Do(reqStale)
	require.NoError(t, err)
	defer respStale.Body.Close()

	assert.Equal(t, http.StatusSeeOther, respStale.StatusCode, "Stale POST should be captured and return 303")
	location := respStale.Header.Get("Location")
		resumeToken := extractResumeToken(location)
	require.NotEmpty(t, resumeToken, "Resume token should be generated")

	client.CheckRedirect = nil
	loginUserFunc(t, serverURL, "alice", "alice-test", client)

	// Resume the action
	reqResumeGet, _ := http.NewRequest("GET", serverURL+"/resume?token="+resumeToken, nil)
	respResumeGet, err := client.Do(reqResumeGet)
	require.NoError(t, err)
	bodyResumeGet, _ := io.ReadAll(respResumeGet.Body)
	respResumeGet.Body.Close()
	docResumeGet, _ := goquery.NewDocumentFromReader(strings.NewReader(string(bodyResumeGet)))
	csrfResume, exists := docResumeGet.Find("input[name='gorilla.csrf.Token']").Attr("value")
	if !exists {
		t.Logf("GET /resume Body: %s", string(bodyResumeGet))
	}
	require.True(t, exists, "CSRF field must exist on resume page")
	require.NotEmpty(t, csrfResume, "CSRF field must not be empty on resume page")

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resumeVals := url.Values{"token": {resumeToken}, "gorilla.csrf.Token": {csrfResume}}
	reqResumeAction, _ := http.NewRequest("POST", serverURL+"/resume", strings.NewReader(resumeVals.Encode()))
	reqResumeAction.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqResumeAction.Header.Set("Sec-Fetch-Site", "same-origin")
	respResumeAction, err := client.Do(reqResumeAction)
	require.NoError(t, err)
	bodyResumeAction, _ := io.ReadAll(respResumeAction.Body)
	respResumeAction.Body.Close()

	assert.Equal(t, http.StatusOK, respResumeAction.StatusCode)
	assert.NotEmpty(t, string(bodyResumeAction), "Success body should not be empty")

	countAfter, _ := dbProbe.AdminCountForumTopics(context.Background())
	assert.Equal(t, countBefore+1, countAfter, "Exactly one topic should be created after resume")

	var latestTitle string
	_ = srv.DB.QueryRow("SELECT title FROM forumtopic ORDER BY idforumtopic DESC LIMIT 1").Scan(&latestTitle)
	assert.Equal(t, "Bob Topic", latestTitle)

	// Resume again should fail
	reqResumeAgain, _ := http.NewRequest("POST", serverURL+"/resume", strings.NewReader(resumeVals.Encode()))
	reqResumeAgain.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqResumeAgain.Header.Set("Sec-Fetch-Site", "same-origin")
	respResumeAgain, err := client.Do(reqResumeAgain)
	require.NoError(t, err)
	respResumeAgain.Body.Close()
	assert.Equal(t, http.StatusNotFound, respResumeAgain.StatusCode)

	countAfter2, _ := dbProbe.AdminCountForumTopics(context.Background())
	assert.Equal(t, countAfter, countAfter2, "Topic should not be created twice")
}

// 3. Authorization revoked then restored
func TestResume_RevokedThenRestored(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	_ = srv
	serverURL := httpServer.URL
	dbProbe := srv.Queries
	_ = srv

	client := createClient()

	loginUserFunc(t, serverURL, "alice", "alice-test", client)

	countBefore, _ := dbProbe.AdminCountForumTopics(context.Background())

	reqGet, _ := http.NewRequest("GET", serverURL+"/private/topic/new", nil)
	respGet, _ := client.Do(reqGet)
	bodyGet, _ := io.ReadAll(respGet.Body)
	respGet.Body.Close()
	nonce := extractNonce(string(bodyGet))

		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(bodyGet)))
	formAction, exists := doc.Find("form#private-form").Attr("action")
	require.True(t, exists, "Form action attribute must exist")
	require.Equal(t, "/private/topic/new", formAction, "Form action must be exactly /private/topic/new")
	parsedAction, _ := url.Parse(formAction)
	actionURL := reqGet.URL.ResolveReference(parsedAction)

	// True logout
	reqLogoutGet, _ := http.NewRequest("GET", serverURL+"/usr/logout", nil)
	respLogoutGet, _ := client.Do(reqLogoutGet)
	bodyLogoutGet, _ := io.ReadAll(respLogoutGet.Body)
	respLogoutGet.Body.Close()

	require.Equal(t, 200, respLogoutGet.StatusCode)


	docLogout, _ := goquery.NewDocumentFromReader(strings.NewReader(string(bodyLogoutGet)))
	logoutCsrfField, _ := docLogout.Find("form[action='/usr/logout'] input[name='gorilla.csrf.Token']").Attr("value")
			require.NotEmpty(t, logoutCsrfField)

	logoutForm := url.Values{}
	logoutForm.Add("gorilla.csrf.Token", logoutCsrfField)
	reqLogoutPost, _ := http.NewRequest("POST", serverURL+"/usr/logout", strings.NewReader(logoutForm.Encode()))
	reqLogoutPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqLogoutPost.Header.Set("X-CSRF-Token", logoutCsrfField)
	reqLogoutPost.Header.Set("Referer", serverURL+"/usr/logout")
	reqLogoutPost.Header.Set("X-CSRF-Token", logoutCsrfField)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	respLogoutPost, _ := client.Do(reqLogoutPost)
	respLogoutPost.Body.Close()

	formStale := url.Values{
		"task":               {"Private topic create"},
		"participants":       {"bob"},
		"title":              {"Bob Topic Revoked"},
		"description":        {"Bob test"},
		"gorilla.csrf.Token": {"stale-csrf-token"},
		"resume_nonce":       {nonce},
	}
	reqStale, _ := http.NewRequest("POST", actionURL.String(), strings.NewReader(formStale.Encode()))
	reqStale.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respStale, _ := client.Do(reqStale)
	location := respStale.Header.Get("Location")
		resumeToken := extractResumeToken(location)
	respStale.Body.Close()

	// Revoke auth
	_, err := srv.DB.Exec("DELETE FROM grants WHERE user_id=1 AND section='privateforum' AND item='topic'")
	require.NoError(t, err)

	client.CheckRedirect = nil
	loginUserFunc(t, serverURL, "alice", "alice-test", client)

	reqResumeGet, _ := http.NewRequest("GET", serverURL+"/resume?token="+resumeToken, nil)
	respResumeGet, _ := client.Do(reqResumeGet)
	bodyResumeGet, _ := io.ReadAll(respResumeGet.Body)
	respResumeGet.Body.Close()
	docResumeGet, _ := goquery.NewDocumentFromReader(strings.NewReader(string(bodyResumeGet)))
	csrfResume, exists := docResumeGet.Find("input[name='gorilla.csrf.Token']").Attr("value")
	if !exists {
		t.Logf("GET /resume Body: %s", string(bodyResumeGet))
	}
	require.True(t, exists, "CSRF field must exist on resume page")
	require.NotEmpty(t, csrfResume, "CSRF field must not be empty on resume page")

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }

	// Attempt resume - should fail 403
	resumeVals := url.Values{"token": {resumeToken}, "gorilla.csrf.Token": {csrfResume}}
	reqResumeAction, _ := http.NewRequest("POST", serverURL+"/resume", strings.NewReader(resumeVals.Encode()))
	reqResumeAction.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqResumeAction.Header.Set("Sec-Fetch-Site", "same-origin")
	respResumeAction, _ := client.Do(reqResumeAction)
	respResumeAction.Body.Close()
	assert.Equal(t, http.StatusForbidden, respResumeAction.StatusCode, "Revoked user should get 403")

	countDenied, _ := dbProbe.AdminCountForumTopics(context.Background())
	assert.Equal(t, countBefore, countDenied, "No new topic should be created on denial")

	// Restore auth
	_, err = srv.DB.Exec("INSERT INTO grants (user_id, section, item, rule_type, action, active) VALUES (1, 'privateforum', 'topic', 'allow', 'see', 1)")
	require.NoError(t, err)
	_, err = srv.DB.Exec("INSERT INTO grants (user_id, section, item, rule_type, action, active) VALUES (1, 'privateforum', 'topic', 'allow', 'create', 1)")
	require.NoError(t, err)

	// Retry
	reqResumeActionRestored, _ := http.NewRequest("POST", serverURL+"/resume", strings.NewReader(resumeVals.Encode()))
	reqResumeActionRestored.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqResumeActionRestored.Header.Set("Sec-Fetch-Site", "same-origin")
	respResumeActionRestored, _ := client.Do(reqResumeActionRestored)
	respResumeActionRestored.Body.Close()
	assert.Equal(t, http.StatusOK, respResumeActionRestored.StatusCode, "Restored user should succeed")

	countRestored, _ := dbProbe.AdminCountForumTopics(context.Background())
	assert.Equal(t, countBefore+1, countRestored, "Topic should be created after restoration")
}

// 4. Validation failure / at-most-once
func TestResume_ValidationFailure(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	_ = srv
	serverURL := httpServer.URL
	dbProbe := srv.Queries
	_ = srv

	client := createClient()
	loginUserFunc(t, serverURL, "alice", "alice-test", client)

	countBefore, _ := dbProbe.AdminCountForumTopics(context.Background())

	reqGet, _ := http.NewRequest("GET", serverURL+"/private/topic/new", nil)
	respGet, _ := client.Do(reqGet)
	bodyGet, _ := io.ReadAll(respGet.Body)
	respGet.Body.Close()
	nonce := extractNonce(string(bodyGet))
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(bodyGet)))
	formAction, exists := doc.Find("form#private-form").Attr("action")
	require.True(t, exists, "Form action attribute must exist")
	require.Equal(t, "/private/topic/new", formAction, "Form action must be exactly /private/topic/new")
	parsedAction, _ := url.Parse(formAction)
	actionURL := reqGet.URL.ResolveReference(parsedAction)

	reqLogoutGet, _ := http.NewRequest("GET", serverURL+"/usr/logout", nil)
	respLogoutGet, _ := client.Do(reqLogoutGet)
	bodyLogoutGet, _ := io.ReadAll(respLogoutGet.Body)
	respLogoutGet.Body.Close()

	require.Equal(t, 200, respLogoutGet.StatusCode)


	docLogout, _ := goquery.NewDocumentFromReader(strings.NewReader(string(bodyLogoutGet)))
	logoutCsrfField, _ := docLogout.Find("form[action='/usr/logout'] input[name='gorilla.csrf.Token']").Attr("value")
			require.NotEmpty(t, logoutCsrfField)

	logoutForm := url.Values{}
	logoutForm.Add("gorilla.csrf.Token", logoutCsrfField)
	reqLogoutPost, _ := http.NewRequest("POST", serverURL+"/usr/logout", strings.NewReader(logoutForm.Encode()))
	reqLogoutPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqLogoutPost.Header.Set("X-CSRF-Token", logoutCsrfField)
	reqLogoutPost.Header.Set("Referer", serverURL+"/usr/logout")
	reqLogoutPost.Header.Set("X-CSRF-Token", logoutCsrfField)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	respLogoutPost, _ := client.Do(reqLogoutPost)
	respLogoutPost.Body.Close()

	// Validation failure: invalid participants
	formStale := url.Values{
		"task":               {"Private topic create"},
		"participants":       {"non_existent_xyz123"},
		"title":              {"Validation fail"},
		"description":        {"Test"},
		"gorilla.csrf.Token": {"stale-csrf-token"},
		"resume_nonce":       {nonce},
	}
	reqStale, _ := http.NewRequest("POST", actionURL.String(), strings.NewReader(formStale.Encode()))
	reqStale.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respStale, _ := client.Do(reqStale)
	location := respStale.Header.Get("Location")
		resumeToken := extractResumeToken(location)
	respStale.Body.Close()

	client.CheckRedirect = nil
	loginUserFunc(t, serverURL, "alice", "alice-test", client)

	reqResumeGet, _ := http.NewRequest("GET", serverURL+"/resume?token="+resumeToken, nil)
	respResumeGet, _ := client.Do(reqResumeGet)
	bodyResumeGet, _ := io.ReadAll(respResumeGet.Body)
	respResumeGet.Body.Close()
	docResumeGet, _ := goquery.NewDocumentFromReader(strings.NewReader(string(bodyResumeGet)))
	csrfResume, exists := docResumeGet.Find("input[name='gorilla.csrf.Token']").Attr("value")
	if !exists {
		t.Logf("GET /resume Body: %s", string(bodyResumeGet))
	}
	require.True(t, exists, "CSRF field must exist on resume page")
	require.NotEmpty(t, csrfResume, "CSRF field must not be empty on resume page")

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	resumeVals := url.Values{"token": {resumeToken}, "gorilla.csrf.Token": {csrfResume}}
	reqResumeAction, _ := http.NewRequest("POST", serverURL+"/resume", strings.NewReader(resumeVals.Encode()))
	reqResumeAction.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqResumeAction.Header.Set("Sec-Fetch-Site", "same-origin")
	respResumeAction, _ := client.Do(reqResumeAction)
	bodyResumeAction, _ := io.ReadAll(respResumeAction.Body)
	respResumeAction.Body.Close()

	assert.Equal(t, http.StatusOK, respResumeAction.StatusCode, "Validation failure should return 200 OK rendering the form")
	assert.Contains(t, string(bodyResumeAction), "Invalid users: non_existent_xyz123")

	countValidationFail, _ := dbProbe.AdminCountForumTopics(context.Background())
	assert.Equal(t, countBefore, countValidationFail, "No topic should be created on validation fail")

	reqResumeRetry, _ := http.NewRequest("POST", serverURL+"/resume", strings.NewReader(resumeVals.Encode()))
	reqResumeRetry.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqResumeRetry.Header.Set("Sec-Fetch-Site", "same-origin")
	respResumeRetry, _ := client.Do(reqResumeRetry)
	respResumeRetry.Body.Close()
	assert.Equal(t, http.StatusNotFound, respResumeRetry.StatusCode, "Token should be burned")
}

// 5. Mismatched task / wrong user
func TestResume_MismatchedTask(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	_ = srv
	serverURL := httpServer.URL

	client := createClient()
	loginUserFunc(t, serverURL, "bob", "bob-test", client)

	reqGet, _ := http.NewRequest("GET", serverURL+"/private/topic/new", nil)
	respGet, _ := client.Do(reqGet)
	bodyGet, _ := io.ReadAll(respGet.Body)
	respGet.Body.Close()
	nonce := extractNonce(string(bodyGet))
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(bodyGet)))
	formAction, exists := doc.Find("form#private-form").Attr("action")
	require.True(t, exists, "Form action attribute must exist")
	require.Equal(t, "/private/topic/new", formAction, "Form action must be exactly /private/topic/new")
	parsedAction, _ := url.Parse(formAction)
	actionURL := reqGet.URL.ResolveReference(parsedAction)

	reqLogoutGet, _ := http.NewRequest("GET", serverURL+"/usr/logout", nil)
	respLogoutGet, _ := client.Do(reqLogoutGet)
	bodyLogoutGet, _ := io.ReadAll(respLogoutGet.Body)
	respLogoutGet.Body.Close()

	require.Equal(t, 200, respLogoutGet.StatusCode)


	docLogout, _ := goquery.NewDocumentFromReader(strings.NewReader(string(bodyLogoutGet)))
	logoutCsrfField, _ := docLogout.Find("form[action='/usr/logout'] input[name='gorilla.csrf.Token']").Attr("value")
			require.NotEmpty(t, logoutCsrfField)

	logoutForm := url.Values{}
	logoutForm.Add("gorilla.csrf.Token", logoutCsrfField)
	reqLogoutPost, _ := http.NewRequest("POST", serverURL+"/usr/logout", strings.NewReader(logoutForm.Encode()))
	reqLogoutPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqLogoutPost.Header.Set("X-CSRF-Token", logoutCsrfField)
	reqLogoutPost.Header.Set("Referer", serverURL+"/usr/logout")
	reqLogoutPost.Header.Set("X-CSRF-Token", logoutCsrfField)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	respLogoutPost, _ := client.Do(reqLogoutPost)
	respLogoutPost.Body.Close()

	// Task mismatch
	formMismatched := url.Values{
		"task":               {"fakeTaskCreate"},
		"participants":       {"bob"},
		"gorilla.csrf.Token": {"stale-csrf-token"},
		"resume_nonce":       {nonce},
	}
	reqStaleMismatch, _ := http.NewRequest("POST", actionURL.String(), strings.NewReader(formMismatched.Encode()))
	reqStaleMismatch.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	respStaleMismatch, _ := client.Do(reqStaleMismatch)
	respStaleMismatch.Body.Close()
	assert.Equal(t, http.StatusForbidden, respStaleMismatch.StatusCode, "Mismatched task should be rejected")

	// Wrong user
	clientBob := createClient()
	loginUserFunc(t, serverURL, "bob", "bob-test", clientBob)
	reqGetBob, _ := http.NewRequest("GET", serverURL+"/private/topic/new", nil)
	respGetBob, _ := clientBob.Do(reqGetBob)
	bodyGetBob, _ := io.ReadAll(respGetBob.Body)
	respGetBob.Body.Close()
	// nonceBob := extractNonce(string(bodyGetBob))
	docBob, _ := goquery.NewDocumentFromReader(strings.NewReader(string(bodyGetBob)))
	formActionBob, existsBob := docBob.Find("form#private-form").Attr("action")
	require.True(t, existsBob)
	parsedActionBob, _ := url.Parse(formActionBob)
	actionURLBob := reqGetBob.URL.ResolveReference(parsedActionBob)

	// Bob steals Alice's nonce
	formStaleWrongUser := url.Values{
		"task":               {"Private topic create"},
		"participants":       {"alice"},
		"gorilla.csrf.Token": {"stale-csrf-token"},
		"resume_nonce":       {nonce}, // Alice's nonce
	}

	// Bob logs out
	reqLogoutGetBob, _ := http.NewRequest("GET", serverURL+"/usr/logout", nil)
	respLogoutGetBob, _ := clientBob.Do(reqLogoutGetBob)
	bodyLogoutGetBob, _ := io.ReadAll(respLogoutGetBob.Body)
	respLogoutGetBob.Body.Close()
	docLogoutBob, _ := goquery.NewDocumentFromReader(strings.NewReader(string(bodyLogoutGetBob)))
	logoutCsrfFieldBob, _ := docLogoutBob.Find("form[action='/usr/logout'] input[name='gorilla.csrf.Token']").Attr("value")

	logoutFormBob := url.Values{}
	logoutFormBob.Add("gorilla.csrf.Token", logoutCsrfFieldBob)
	reqLogoutPostBob, _ := http.NewRequest("POST", serverURL+"/usr/logout", strings.NewReader(logoutFormBob.Encode()))
	reqLogoutPostBob.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqLogoutPostBob.Header.Set("X-CSRF-Token", logoutCsrfFieldBob)
	clientBob.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	respLogoutPostBob, _ := clientBob.Do(reqLogoutPostBob)
	respLogoutPostBob.Body.Close()

	reqStaleWrongUser, _ := http.NewRequest("POST", actionURLBob.String(), strings.NewReader(formStaleWrongUser.Encode()))
	reqStaleWrongUser.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	respStaleWrongUser, _ := clientBob.Do(reqStaleWrongUser)
	respStaleWrongUser.Body.Close()
	assert.Equal(t, http.StatusForbidden, respStaleWrongUser.StatusCode, "Wrong user stealing nonce should be rejected")
}
