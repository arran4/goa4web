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

	"github.com/arran4/goa4web/internal/app/server"
	"github.com/arran4/goa4web/testdata/scenarios"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func loginUserFunc(t *testing.T, serverURL, username, password string, client *http.Client) {
	reqGet, err := http.NewRequest(http.MethodGet, serverURL+"/login", nil)
	require.NoError(t, err)
	respGet, err := client.Do(reqGet)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, respGet.StatusCode)
	body, err := io.ReadAll(respGet.Body)
	require.NoError(t, err)
	respGet.Body.Close()

	csrfField := inputValue(string(body), "gorilla.csrf.Token")
	require.NotEmpty(t, csrfField)

	form := url.Values{}
	form.Add("username", username)
	form.Add("password", password)
	form.Add("gorilla.csrf.Token", csrfField)
	form.Add("task", "Login")

	reqPost, err := http.NewRequest(http.MethodPost, serverURL+"/login", strings.NewReader(form.Encode()))
	require.NoError(t, err)
	reqPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	oldRedirect := client.CheckRedirect
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	respPost, err := client.Do(reqPost)
	require.NoError(t, err)
	bodyPost, err := io.ReadAll(respPost.Body)
	require.NoError(t, err)
	respPost.Body.Close()
	require.Equal(t, http.StatusSeeOther, respPost.StatusCode, "successful login must redirect")
	require.NotContains(t, string(bodyPost), "Invalid username or password")

	// Prove that the same cookie jar now represents an authenticated session.
	// Keep redirects disabled so an unauthenticated redirect to /login cannot
	// masquerade as a successful 200 response.
	reqVerify, err := http.NewRequest(http.MethodGet, serverURL+"/usr/lang", nil)
	require.NoError(t, err)
	respVerify, err := client.Do(reqVerify)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, respVerify.StatusCode, "login must establish an authenticated session")
	respVerify.Body.Close()

	client.CheckRedirect = oldRedirect
}

func extractNonce(body string) string {
	return inputValue(body, "resume_nonce")
}

func parseHTML(body string) *html.Node {
	root, _ := html.Parse(strings.NewReader(body))
	return root
}

func nodeAttribute(node *html.Node, name string) (string, bool) {
	if node == nil {
		return "", false
	}
	for _, attr := range node.Attr {
		if attr.Key == name {
			return attr.Val, true
		}
	}
	return "", false
}

func findElement(root *html.Node, tag string, attributes map[string]string) *html.Node {
	if root == nil {
		return nil
	}
	if root.Type == html.ElementNode && root.Data == tag {
		matches := true
		for name, want := range attributes {
			got, ok := nodeAttribute(root, name)
			if !ok || got != want {
				matches = false
				break
			}
		}
		if matches {
			return root
		}
	}
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		if found := findElement(child, tag, attributes); found != nil {
			return found
		}
	}
	return nil
}

func inputValue(body, name string) string {
	input := findElement(parseHTML(body), "input", map[string]string{"name": name})
	value, _ := nodeAttribute(input, "value")
	return value
}

func formAction(body, id string) string {
	form := findElement(parseHTML(body), "form", map[string]string{"id": id})
	action, _ := nodeAttribute(form, "action")
	return action
}

func logoutUserFunc(t *testing.T, serverURL, csrfToken string, client *http.Client) *http.Response {
	t.Helper()
	require.NotEmpty(t, csrfToken)
	form := url.Values{"gorilla.csrf.Token": {csrfToken}}
	req, err := http.NewRequest(http.MethodPost, serverURL+"/usr/logout", strings.NewReader(form.Encode()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.Header.Set("Referer", serverURL+"/usr/logout")
	oldRedirect := client.CheckRedirect
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, resp.StatusCode)
	client.CheckRedirect = oldRedirect
	return resp
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
	respGet, err := client.Do(reqGet)
	require.NoError(t, err)
	bodyGet, _ := io.ReadAll(respGet.Body)
	respGet.Body.Close()

	nonce := extractNonce(string(bodyGet))
	if nonce == "" {
		t.Logf("Failed to get nonce. Body: %s", string(bodyGet))
	}
	require.NotEmpty(t, nonce)

	formAction := formAction(string(bodyGet), "private-form")
	require.NotEmpty(t, formAction, "Form action attribute must exist")
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
	assert.Empty(t, respStale.Header.Get("Location"), "bad CSRF must not generate a login continuation")
	requireNoStore(t, respStale)

	// Verify no resumable payload was captured.
	var count int
	_ = srv.DB.QueryRow("SELECT COUNT(*) FROM pending_actions").Scan(&count)
	assert.Equal(t, 1, count, "No new pending action should be stored for authenticated user, only the one from GET")
	var formData string
	require.NoError(t, srv.DB.QueryRow("SELECT form_data FROM pending_actions WHERE id = ?", hashOpaque(nonce)).Scan(&formData))
	assert.Empty(t, formData, "authenticated bad CSRF must leave the form nonce uncaptured")
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
	respGet, err := client.Do(reqGet)
	require.NoError(t, err)
	bodyGet, _ := io.ReadAll(respGet.Body)
	respGet.Body.Close()
	nonce := extractNonce(string(bodyGet))

	formAction := formAction(string(bodyGet), "private-form")
	require.NotEmpty(t, formAction, "Form action attribute must exist")
	require.Equal(t, "/private/topic/new", formAction, "Form action must be exactly /private/topic/new")
	parsedAction, _ := url.Parse(formAction)
	actionURL := reqGet.URL.ResolveReference(parsedAction)

	respLogout := logoutUserFunc(t, serverURL, inputValue(string(bodyGet), "gorilla.csrf.Token"), client)
	respLogout.Body.Close()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }

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

	// Assert the pending row still exists and is unconsumed
	var consumedAt *string
	err = srv.DB.QueryRow("SELECT consumed_at FROM pending_actions").Scan(&consumedAt)
	require.NoError(t, err)
	require.Nil(t, consumedAt, "GET must never consume the pending action")

	csrfResume := inputValue(string(bodyResumeGet), "gorilla.csrf.Token")
	if csrfResume == "" {
		t.Logf("GET /resume Body: %s", string(bodyResumeGet))
	}
	require.NotEmpty(t, csrfResume, "CSRF field must not be empty on resume page")

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resumeVals := url.Values{"token": {resumeToken}, "operation": {"resume"}, "gorilla.csrf.Token": {csrfResume}}
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

	// Assert the pending row was consumed
	var consumedAtAfter *string
	err = srv.DB.QueryRow("SELECT consumed_at FROM pending_actions").Scan(&consumedAtAfter)
	require.NoError(t, err)
	require.NotNil(t, consumedAtAfter, "POST must consume the pending action")

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
	respGet, err := client.Do(reqGet)
	require.NoError(t, err)
	bodyGet, _ := io.ReadAll(respGet.Body)
	respGet.Body.Close()
	nonce := extractNonce(string(bodyGet))

	formAction := formAction(string(bodyGet), "private-form")
	require.NotEmpty(t, formAction, "Form action attribute must exist")
	require.Equal(t, "/private/topic/new", formAction, "Form action must be exactly /private/topic/new")
	parsedAction, _ := url.Parse(formAction)
	actionURL := reqGet.URL.ResolveReference(parsedAction)

	respLogout := logoutUserFunc(t, serverURL, inputValue(string(bodyGet), "gorilla.csrf.Token"), client)
	respLogout.Body.Close()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }

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
	_, err = srv.DB.Exec("DELETE FROM grants WHERE user_id=1 AND section='privateforum' AND item='topic'")
	require.NoError(t, err)

	client.CheckRedirect = nil
	loginUserFunc(t, serverURL, "alice", "alice-test", client)

	reqResumeGet, _ := http.NewRequest("GET", serverURL+"/resume?token="+resumeToken, nil)
	respResumeGet, _ := client.Do(reqResumeGet)
	bodyResumeGet, _ := io.ReadAll(respResumeGet.Body)
	respResumeGet.Body.Close()
	csrfResume := inputValue(string(bodyResumeGet), "gorilla.csrf.Token")
	if csrfResume == "" {
		t.Logf("GET /resume Body: %s", string(bodyResumeGet))
	}
	require.NotEmpty(t, csrfResume, "CSRF field must not be empty on resume page")

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }

	// Attempt resume - should fail 403
	resumeVals := url.Values{"token": {resumeToken}, "operation": {"resume"}, "gorilla.csrf.Token": {csrfResume}}
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
	respGet, err := client.Do(reqGet)
	require.NoError(t, err)
	bodyGet, _ := io.ReadAll(respGet.Body)
	respGet.Body.Close()
	nonce := extractNonce(string(bodyGet))
	formAction := formAction(string(bodyGet), "private-form")
	require.NotEmpty(t, formAction, "Form action attribute must exist")
	require.Equal(t, "/private/topic/new", formAction, "Form action must be exactly /private/topic/new")
	parsedAction, _ := url.Parse(formAction)
	actionURL := reqGet.URL.ResolveReference(parsedAction)

	respLogout := logoutUserFunc(t, serverURL, inputValue(string(bodyGet), "gorilla.csrf.Token"), client)
	respLogout.Body.Close()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }

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
	csrfResume := inputValue(string(bodyResumeGet), "gorilla.csrf.Token")
	if csrfResume == "" {
		t.Logf("GET /resume Body: %s", string(bodyResumeGet))
	}
	require.NotEmpty(t, csrfResume, "CSRF field must not be empty on resume page")

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	resumeVals := url.Values{"token": {resumeToken}, "operation": {"resume"}, "gorilla.csrf.Token": {csrfResume}}
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

// 5. Mismatched task
func TestResume_MismatchedTask(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	_ = srv
	serverURL := httpServer.URL

	client := createClient()
	loginUserFunc(t, serverURL, "alice", "alice-test", client)

	reqGet, _ := http.NewRequest("GET", serverURL+"/private/topic/new", nil)
	respGet, err := client.Do(reqGet)
	require.NoError(t, err)
	bodyGet, _ := io.ReadAll(respGet.Body)
	respGet.Body.Close()
	nonce := extractNonce(string(bodyGet))

	formAction := formAction(string(bodyGet), "private-form")
	require.NotEmpty(t, formAction, "Form action attribute must exist")
	require.Equal(t, "/private/topic/new", formAction, "Form action must be exactly /private/topic/new")
	parsedAction, _ := url.Parse(formAction)
	actionURL := reqGet.URL.ResolveReference(parsedAction)

	respLogout := logoutUserFunc(t, serverURL, inputValue(string(bodyGet), "gorilla.csrf.Token"), client)
	respLogout.Body.Close()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }

	// DB Action Type Mismatch
	res, err := srv.DB.Exec("UPDATE pending_actions SET action_type='invalidType' WHERE form_data=''")
	require.NoError(t, err)
	rows, _ := res.RowsAffected()
	require.Greater(t, rows, int64(0), "UPDATE pending_actions must affect at least 1 row")

	formStaleDbmismatch := url.Values{
		"task":               {"Private topic create"},
		"participants":       {"bob"},
		"gorilla.csrf.Token": {"stale-csrf-token"},
		"resume_nonce":       {nonce},
	}
	reqStaleDbMismatch, _ := http.NewRequest("POST", actionURL.String(), strings.NewReader(formStaleDbmismatch.Encode()))
	reqStaleDbMismatch.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respStaleDbMismatch, err := client.Do(reqStaleDbMismatch)
	require.NoError(t, err)
	respStaleDbMismatch.Body.Close()
	assert.Equal(t, http.StatusForbidden, respStaleDbMismatch.StatusCode, "Mismatched DB action type must be rejected")

	// Restore DB action type
	_, err = srv.DB.Exec("UPDATE pending_actions SET action_type='Private topic create' WHERE form_data=''")
	require.NoError(t, err)

	// Task mismatch
	formMismatched := url.Values{
		"task":               {"fakeTaskCreate"},
		"participants":       {"alice"},
		"gorilla.csrf.Token": {"stale-csrf-token"},
		"resume_nonce":       {nonce},
	}
	reqStaleMismatch, _ := http.NewRequest("POST", actionURL.String(), strings.NewReader(formMismatched.Encode()))
	reqStaleMismatch.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	respStaleMismatch, err := client.Do(reqStaleMismatch)
	require.NoError(t, err)
	respStaleMismatch.Body.Close()
	assert.Equal(t, http.StatusForbidden, respStaleMismatch.StatusCode, "Mismatched task should be rejected")
}

func TestResume_SameBrowserWrongUser(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	serverURL := httpServer.URL
	client := createClient()

	loginUserFunc(t, serverURL, "alice", "alice-test", client)

	countBefore, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)

	// Alice renders the resumable form.
	reqGet, err := http.NewRequest(http.MethodGet, serverURL+"/private/topic/new", nil)
	require.NoError(t, err)
	respGet, err := client.Do(reqGet)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, respGet.StatusCode)
	bodyGet, err := io.ReadAll(respGet.Body)
	require.NoError(t, err)
	respGet.Body.Close()

	formAction := formAction(string(bodyGet), "private-form")
	require.NotEmpty(t, formAction)
	require.Equal(t, "/private/topic/new", formAction)
	nonce := extractNonce(string(bodyGet))
	require.NotEmpty(t, nonce)
	parsedAction, err := url.Parse(formAction)
	require.NoError(t, err)
	actionURL := reqGet.URL.ResolveReference(parsedAction)

	respLogout := logoutUserFunc(t, serverURL, inputValue(string(bodyGet), "gorilla.csrf.Token"), client)
	respLogout.Body.Close()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }

	// Alice's already-rendered stale form is captured while unauthenticated.
	formStale := url.Values{
		"task":               {"Private topic create"},
		"participants":       {"bob"},
		"title":              {"Account switch must not replay"},
		"description":        {"same browser"},
		"gorilla.csrf.Token": {"stale-csrf-token"},
		"resume_nonce":       {nonce},
	}
	reqStale, err := http.NewRequest(http.MethodPost, actionURL.String(), strings.NewReader(formStale.Encode()))
	require.NoError(t, err)
	reqStale.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respStale, err := client.Do(reqStale)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, respStale.StatusCode)
	resumeToken := extractResumeToken(respStale.Header.Get("Location"))
	respStale.Body.Close()
	require.NotEmpty(t, resumeToken)

	tokenHash := sha256.Sum256([]byte(resumeToken))
	tokenHashHex := hex.EncodeToString(tokenHash[:])

	// Bob logs into the same browser/cookie jar: browser ID is unchanged.
	client.CheckRedirect = nil
	loginUserFunc(t, serverURL, "bob", "bob-test", client)

	reqResume, err := http.NewRequest(http.MethodGet, serverURL+"/resume?token="+url.QueryEscape(resumeToken), nil)
	require.NoError(t, err)
	respResume, err := client.Do(reqResume)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, respResume.StatusCode, "different UID in same browser must not resume Alice's action")
	respResume.Body.Close()

	countAfter, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)
	require.Equal(t, countBefore, countAfter, "identity denial must create no topic")

	var consumedAt *string
	err = srv.DB.QueryRow("SELECT consumed_at FROM pending_actions WHERE id = ?", tokenHashHex).Scan(&consumedAt)
	require.NoError(t, err)
	require.Nil(t, consumedAt, "identity denial must not consume Alice's token")
}
