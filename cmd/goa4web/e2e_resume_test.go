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
	"regexp"
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
	// Fetch login page to get CSRF token
	reqLoginGet, _ := http.NewRequest("GET", serverURL+"/login", nil)
	respLoginGet, err := clientA.Do(reqLoginGet)
	require.NoError(t, err)
	loginBody, _ := io.ReadAll(respLoginGet.Body)
	respLoginGet.Body.Close()

	logoutCsrfField := ""
	csrfRegex := regexp.MustCompile(`name="gorilla\.csrf\.Token"[^>]*value="([^"]+)"`)
	matches := csrfRegex.FindStringSubmatch(string(loginBody))
	if len(matches) > 1 {
		logoutCsrfField = matches[1]
	}

	require.NotEmpty(t, logoutCsrfField, "CSRF field should be present")

	logoutForm := url.Values{}
	logoutForm.Add("gorilla.csrf.Token", logoutCsrfField)
	reqLogoutPost, _ := http.NewRequest("POST", serverURL+"/usr/logout", strings.NewReader(logoutForm.Encode()))
	reqLogoutPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respLogoutPost, err := clientA.Do(reqLogoutPost)
	require.NoError(t, err)
	respLogoutPost.Body.Close()

	// Verify old auth is genuinely unusable using the precise target /usr which gives 403 or redirects to login
	reqCheck, _ := http.NewRequest("GET", serverURL+"/usr", nil)

	// Ensure we skip TLS verification for the local test server in this custom client
	noRedirectClient := &http.Client{
		Jar: clientA.Jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	respCheck, err := noRedirectClient.Do(reqCheck)
	require.NoError(t, err)
	respCheck.Body.Close()
	// /usr redirects to /login if not authenticated
	require.Equal(t, http.StatusSeeOther, respCheck.StatusCode)
	require.Contains(t, respCheck.Header.Get("Location"), "/login")

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

	assert.Equal(t, http.StatusSeeOther, respResumeAction.StatusCode) // It actually follows redirects, so it should be 200, wait, our previous change was catching 303.

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
