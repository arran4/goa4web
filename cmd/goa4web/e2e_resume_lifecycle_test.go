//go:build sqlite

package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/arran4/goa4web/core"
	authhandlers "github.com/arran4/goa4web/handlers/auth"
	privateforumhandlers "github.com/arran4/goa4web/handlers/privateforum"
	"github.com/arran4/goa4web/internal/app"
	"github.com/arran4/goa4web/internal/eventbus"
	routerpkg "github.com/arran4/goa4web/internal/router"
	"github.com/stretchr/testify/require"
)

type resumableForm struct {
	nonce string
	csrf  string
}

func fetchResumableForm(t *testing.T, serverURL string, client *http.Client) resumableForm {
	t.Helper()
	resp, err := client.Get(serverURL + "/private/topic/new")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "/private/topic/new", formAction(string(body), "private-form"))
	form := resumableForm{
		nonce: extractNonce(string(body)),
		csrf:  inputValue(string(body), "gorilla.csrf.Token"),
	}
	require.NotEmpty(t, form.nonce)
	require.NotEmpty(t, form.csrf)
	return form
}

func privateTopicValues(form resumableForm, title, participants string) url.Values {
	return url.Values{
		"task":               {"Private topic create"},
		"participants":       {participants},
		"title":              {title},
		"description":        {"resume lifecycle regression"},
		"gorilla.csrf.Token": {form.csrf},
		"resume_nonce":       {form.nonce},
	}
}

func postEncoded(t *testing.T, client *http.Client, target string, values url.Values) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, target, strings.NewReader(values.Encode()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func requireNoStore(t *testing.T, resp *http.Response) {
	t.Helper()
	require.Contains(t, resp.Header.Get("Cache-Control"), "no-store")
	require.Equal(t, "no-store", resp.Header.Get("Cloudflare-CDN-Cache-Control"))
}

func hashOpaque(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

func capturePrivateTopic(t *testing.T, serverURL string, client *http.Client, title, rawQuery string) (string, resumableForm, *http.Response) {
	t.Helper()
	form := fetchResumableForm(t, serverURL, client)
	logoutResp := logoutUserFunc(t, serverURL, form.csrf, client)
	requireNoStore(t, logoutResp)
	logoutResp.Body.Close()

	values := privateTopicValues(form, title, "bob")
	values.Set("gorilla.csrf.Token", "stale-token")
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	resp := postEncoded(t, client, serverURL+"/private/topic/new"+rawQuery, values)
	require.Equal(t, http.StatusSeeOther, resp.StatusCode)
	requireNoStore(t, resp)
	token := extractResumeToken(resp.Header.Get("Location"))
	require.NotEmpty(t, token)
	return token, form, resp
}

func resumeCSRF(t *testing.T, serverURL, token string, client *http.Client) (string, *http.Response) {
	t.Helper()
	resp, err := client.Get(serverURL + "/resume?token=" + url.QueryEscape(token))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()
	csrfToken := inputValue(string(body), "gorilla.csrf.Token")
	require.NotEmpty(t, csrfToken)
	return csrfToken, resp
}

func TestResume_NormalSubmissionRetiresNonceAndValidationRotates(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	client := createClient()
	loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }

	countBefore, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)
	form := fetchResumableForm(t, httpServer.URL, client)
	values := privateTopicValues(form, "normal nonce retirement", "bob")
	resp := postEncoded(t, client, httpServer.URL+"/private/topic/new", values)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	countAfter, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)
	require.Equal(t, countBefore+1, countAfter)

	var consumedAt sql.NullTime
	require.NoError(t, srv.DB.QueryRow("SELECT consumed_at FROM pending_actions WHERE id = ?", hashOpaque(form.nonce)).Scan(&consumedAt))
	require.True(t, consumedAt.Valid)

	logoutResp := logoutUserFunc(t, httpServer.URL, form.csrf, client)
	logoutResp.Body.Close()
	values.Set("gorilla.csrf.Token", "stale-token")
	replay := postEncoded(t, client, httpServer.URL+"/private/topic/new", values)
	require.Equal(t, http.StatusForbidden, replay.StatusCode)
	require.Empty(t, replay.Header.Get("Location"))
	requireNoStore(t, replay)
	replay.Body.Close()
	countAfterReplay, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)
	require.Equal(t, countAfter, countAfterReplay)

	client.CheckRedirect = nil
	loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	invalidForm := fetchResumableForm(t, httpServer.URL, client)
	invalid := privateTopicValues(invalidForm, "correctable form", "missing-user")
	invalidResp := postEncoded(t, client, httpServer.URL+"/private/topic/new", invalid)
	require.Equal(t, http.StatusOK, invalidResp.StatusCode)
	body, err := io.ReadAll(invalidResp.Body)
	require.NoError(t, err)
	invalidResp.Body.Close()
	newNonce := extractNonce(string(body))
	require.NotEmpty(t, newNonce)
	require.NotEqual(t, invalidForm.nonce, newNonce)
	require.Error(t, srv.DB.QueryRow("SELECT id FROM pending_actions WHERE id = ? AND consumed_at IS NULL", hashOpaque(invalidForm.nonce)).Scan(new(string)))
	require.NoError(t, srv.DB.QueryRow("SELECT id FROM pending_actions WHERE id = ? AND consumed_at IS NULL", hashOpaque(newNonce)).Scan(new(string)))
}

func TestResume_ConcurrentNormalFormReuseExecutesOnce(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	client := createClient()
	loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	form := fetchResumableForm(t, httpServer.URL, client)
	values := privateTopicValues(form, "concurrent normal claim", "bob")
	before, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)

	statuses := make(chan int, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, reqErr := http.NewRequest(http.MethodPost, httpServer.URL+"/private/topic/new", strings.NewReader(values.Encode()))
			if reqErr != nil {
				statuses <- 0
				return
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Sec-Fetch-Site", "same-origin")
			resp, doErr := client.Do(req)
			if doErr != nil {
				statuses <- 0
				return
			}
			resp.Body.Close()
			statuses <- resp.StatusCode
		}()
	}
	wg.Wait()
	close(statuses)
	got := []int{}
	for status := range statuses {
		got = append(got, status)
	}
	slices.Sort(got)
	require.Equal(t, []int{http.StatusOK, http.StatusForbidden}, got)
	after, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)
	require.Equal(t, before+1, after)
}

func TestResume_CancellationFreshCSRFAndNoStore(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	client := createClient()
	loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
	before, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)
	token, form, captureResp := capturePrivateTopic(t, httpServer.URL, client, "cancelled pending topic", "?source=stale")
	location := captureResp.Header.Get("Location")
	captureResp.Body.Close()
	require.NotContains(t, location, "cancelled")
	require.NotContains(t, location, form.nonce)
	require.NotContains(t, location, "source")
	loginURL, err := url.Parse(location)
	require.NoError(t, err)
	loginKeys := mapKeys(loginURL.Query())
	slices.Sort(loginKeys)
	require.Equal(t, []string{"back"}, loginKeys)

	loginContinuation, err := client.Get(httpServer.URL + location)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, loginContinuation.StatusCode)
	requireNoStore(t, loginContinuation)
	loginContinuation.Body.Close()

	client.CheckRedirect = nil
	loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	csrfToken, interstitial := resumeCSRF(t, httpServer.URL, token, client)
	requireNoStore(t, interstitial)

	withoutCSRF := postEncoded(t, client, httpServer.URL+"/resume", url.Values{"token": {token}, "operation": {"resume"}})
	require.Equal(t, http.StatusForbidden, withoutCSRF.StatusCode)
	requireNoStore(t, withoutCSRF)
	withoutCSRF.Body.Close()
	var consumed sql.NullTime
	require.NoError(t, srv.DB.QueryRow("SELECT consumed_at FROM pending_actions WHERE id = ?", hashOpaque(token)).Scan(&consumed))
	require.False(t, consumed.Valid)

	var eventMu sync.Mutex
	events := make([]eventbus.TaskEvent, 0, 1)
	srv.Bus.SyncPublish = func(message eventbus.Message) {
		if evt, ok := message.(eventbus.TaskEvent); ok {
			eventMu.Lock()
			events = append(events, evt)
			eventMu.Unlock()
		}
	}
	cancel := postEncoded(t, client, httpServer.URL+"/resume", url.Values{
		"token": {token}, "operation": {"cancel"}, "gorilla.csrf.Token": {csrfToken},
	})
	require.Equal(t, http.StatusSeeOther, cancel.StatusCode)
	require.Equal(t, "/private/topic/new", cancel.Header.Get("Location"))
	requireNoStore(t, cancel)
	cancel.Body.Close()
	eventMu.Lock()
	require.Len(t, events, 1)
	_, isPrivateTopicCreate := events[0].Task.(*privateforumhandlers.PrivateTopicCreateTask)
	eventMu.Unlock()
	require.False(t, isPrivateTopicCreate, "cancellation must not publish as private-topic-create")

	used := postEncoded(t, client, httpServer.URL+"/resume", url.Values{
		"token": {token}, "operation": {"resume"}, "gorilla.csrf.Token": {csrfToken},
	})
	require.Equal(t, http.StatusNotFound, used.StatusCode)
	requireNoStore(t, used)
	used.Body.Close()
	after, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)
	require.Equal(t, before, after)
}

func mapKeys(values url.Values) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func TestResume_ExpiryCleanup(t *testing.T) {
	t.Run("form nonce", func(t *testing.T) {
		httpServer, srv, cleanup := setupTestServer(t)
		defer cleanup()
		client := createClient()
		loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
		form := fetchResumableForm(t, httpServer.URL, client)
		formHash := hashOpaque(form.nonce)
		_, err := srv.DB.Exec("UPDATE pending_actions SET expires_at = ? WHERE id = ?", time.Now().Add(-time.Hour), formHash)
		require.NoError(t, err)
		logoutResp := logoutUserFunc(t, httpServer.URL, form.csrf, client)
		logoutResp.Body.Close()
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
		values := privateTopicValues(form, "expired form", "bob")
		values.Set("gorilla.csrf.Token", "stale-token")
		resp := postEncoded(t, client, httpServer.URL+"/private/topic/new", values)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)
		resp.Body.Close()
		client.CheckRedirect = nil
		loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
		_ = fetchResumableForm(t, httpServer.URL, client)
		var count int
		require.NoError(t, srv.DB.QueryRow("SELECT COUNT(*) FROM pending_actions WHERE id = ?", formHash).Scan(&count))
		require.Zero(t, count)
	})

	t.Run("resume token", func(t *testing.T) {
		httpServer, srv, cleanup := setupTestServer(t)
		defer cleanup()
		client := createClient()
		loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
		token, _, captureResp := capturePrivateTopic(t, httpServer.URL, client, "expired resume", "")
		captureResp.Body.Close()
		tokenHash := hashOpaque(token)
		_, err := srv.DB.Exec("UPDATE pending_actions SET expires_at = ? WHERE id = ?", time.Now().Add(-time.Hour), tokenHash)
		require.NoError(t, err)
		client.CheckRedirect = nil
		loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
		resp, err := client.Get(httpServer.URL + "/resume?token=" + url.QueryEscape(token))
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
		requireNoStore(t, resp)
		resp.Body.Close()
		_ = fetchResumableForm(t, httpServer.URL, client)
		var count int
		require.NoError(t, srv.DB.QueryRow("SELECT COUNT(*) FROM pending_actions WHERE id = ?", tokenHash).Scan(&count))
		require.Zero(t, count)
	})
}

func TestResume_CaptureBoundaries(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	client := createClient()
	loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
	form := fetchResumableForm(t, httpServer.URL, client)
	logoutResp := logoutUserFunc(t, httpServer.URL, form.csrf, client)
	logoutResp.Body.Close()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	values := privateTopicValues(form, "must not capture", "bob")
	values.Set("gorilla.csrf.Token", "stale-token")

	unsupportedReq, err := http.NewRequest(http.MethodPost, httpServer.URL+"/private/topic/new", strings.NewReader(values.Encode()))
	require.NoError(t, err)
	unsupportedReq.Header.Set("Content-Type", "application/json")
	unsupported, err := client.Do(unsupportedReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, unsupported.StatusCode)
	unsupported.Body.Close()

	oversizedValues := url.Values{}
	for key, entries := range values {
		oversizedValues[key] = slices.Clone(entries)
	}
	oversizedValues.Set("description", strings.Repeat("x", 70<<10))
	oversized := postEncoded(t, client, httpServer.URL+"/private/topic/new", oversizedValues)
	require.Equal(t, http.StatusForbidden, oversized.StatusCode)
	oversized.Body.Close()

	nonOptedIn := postEncoded(t, client, httpServer.URL+"/private", values)
	require.Equal(t, http.StatusForbidden, nonOptedIn.StatusCode)
	nonOptedIn.Body.Close()

	var formData string
	require.NoError(t, srv.DB.QueryRow("SELECT form_data FROM pending_actions WHERE id = ?", hashOpaque(form.nonce)).Scan(&formData))
	require.Empty(t, formData)
}

func TestResume_StoredTargetValidation(t *testing.T) {
	for _, target := range []string{"https://example.invalid/private/topic/new", "//example.invalid/private/topic/new", "/private/not-topic-new"} {
		t.Run(target, func(t *testing.T) {
			httpServer, srv, cleanup := setupTestServer(t)
			defer cleanup()
			client := createClient()
			loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
			token, _, captureResp := capturePrivateTopic(t, httpServer.URL, client, "bad stored target", "")
			captureResp.Body.Close()
			var raw string
			require.NoError(t, srv.DB.QueryRow("SELECT form_data FROM pending_actions WHERE id = ?", hashOpaque(token)).Scan(&raw))
			var stored map[string]any
			require.NoError(t, json.Unmarshal([]byte(raw), &stored))
			stored["url"] = target
			tampered, err := json.Marshal(stored)
			require.NoError(t, err)
			_, err = srv.DB.Exec("UPDATE pending_actions SET form_data = ? WHERE id = ?", string(tampered), hashOpaque(token))
			require.NoError(t, err)
			client.CheckRedirect = nil
			loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
			resp, err := client.Get(httpServer.URL + "/resume?token=" + url.QueryEscape(token))
			require.NoError(t, err)
			require.Equal(t, http.StatusForbidden, resp.StatusCode)
			resp.Body.Close()
			var consumed sql.NullTime
			require.NoError(t, srv.DB.QueryRow("SELECT consumed_at FROM pending_actions WHERE id = ?", hashOpaque(token)).Scan(&consumed))
			require.False(t, consumed.Valid)
		})
	}
}

func TestResume_QueryFormPrivacyAndCrossInstanceExecution(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	client := createClient()
	loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
	before, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)
	title := "query and form survive"
	token, form, captureResp := capturePrivateTopic(t, httpServer.URL, client, title, "?source=stale&mode=full")
	location := captureResp.Header.Get("Location")
	captureResp.Body.Close()
	require.NotContains(t, location, title)
	require.NotContains(t, location, form.nonce)
	require.NotContains(t, location, "source")

	var raw string
	require.NoError(t, srv.DB.QueryRow("SELECT form_data FROM pending_actions WHERE id = ?", hashOpaque(token)).Scan(&raw))
	var stored struct {
		Form url.Values `json:"form"`
		URL  string     `json:"url"`
	}
	require.NoError(t, json.Unmarshal([]byte(raw), &stored))
	require.Equal(t, "/private/topic/new?source=stale&mode=full", stored.URL)
	require.Equal(t, title, stored.Form.Get("title"))
	require.Empty(t, stored.Form.Get("resume_nonce"))
	require.Empty(t, stored.Form.Get("gorilla.csrf.Token"))

	baseURL, err := url.Parse(httpServer.URL)
	require.NoError(t, err)
	requestForSession := httptest.NewRequest(http.MethodGet, httpServer.URL+"/", nil)
	for _, cookie := range client.Jar.Cookies(baseURL) {
		requestForSession.AddCookie(cookie)
	}
	session, err := core.Store.Get(requestForSession, core.SessionName)
	require.NoError(t, err)
	sessionState := fmt.Sprint(session.Values)
	require.NotContains(t, sessionState, title)
	require.NotContains(t, sessionState, token)

	client.CheckRedirect = nil
	loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
	secondRouterRegistry := routerpkg.NewRegistry()
	authhandlers.Register(secondRouterRegistry)
	second, err := app.NewServer(context.Background(), srv.Config, nil,
		app.WithStore(srv.Store),
		app.WithDB(srv.DB),
		app.WithQuerier(srv.Queries),
		app.WithDBRegistry(srv.DBReg),
		app.WithEmailRegistry(srv.EmailReg),
		app.WithDLQRegistry(srv.DLQReg),
		app.WithTasksRegistry(srv.TasksReg),
		app.WithRouterRegistry(secondRouterRegistry),
	)
	require.NoError(t, err)
	secondHTTP := httptest.NewTLSServer(second.Router)
	defer secondHTTP.Close()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	csrfToken, interstitial := resumeCSRF(t, secondHTTP.URL, token, client)
	requireNoStore(t, interstitial)
	resume := postEncoded(t, client, secondHTTP.URL+"/resume", url.Values{
		"token": {token}, "operation": {"resume"}, "gorilla.csrf.Token": {csrfToken},
	})
	require.Equal(t, http.StatusOK, resume.StatusCode)
	requireNoStore(t, resume)
	resume.Body.Close()
	after, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)
	require.Equal(t, before+1, after)
}

func TestResume_UnrelatedBrowserDeniedAndTokenUnconsumed(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	original := createClient()
	loginUserFunc(t, httpServer.URL, "alice", "alice-test", original)
	token, _, captureResp := capturePrivateTopic(t, httpServer.URL, original, "wrong browser", "")
	captureResp.Body.Close()

	other := createClient()
	loginUserFunc(t, httpServer.URL, "alice", "alice-test", other)
	other.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := other.Get(httpServer.URL + "/resume?token=" + url.QueryEscape(token))
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
	var consumed sql.NullTime
	require.NoError(t, srv.DB.QueryRow("SELECT consumed_at FROM pending_actions WHERE id = ?", hashOpaque(token)).Scan(&consumed))
	require.False(t, consumed.Valid)
}

func TestResume_ConcurrentResumeTokenExecutesOnce(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	client := createClient()
	loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
	before, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)
	token, _, captureResp := capturePrivateTopic(t, httpServer.URL, client, "concurrent resume claim", "")
	captureResp.Body.Close()
	client.CheckRedirect = nil
	loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	csrfToken, _ := resumeCSRF(t, httpServer.URL, token, client)
	values := url.Values{"token": {token}, "operation": {"resume"}, "gorilla.csrf.Token": {csrfToken}}

	statuses := make(chan int, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, reqErr := http.NewRequest(http.MethodPost, httpServer.URL+"/resume", strings.NewReader(values.Encode()))
			if reqErr != nil {
				statuses <- 0
				return
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Sec-Fetch-Site", "same-origin")
			resp, doErr := client.Do(req)
			if doErr != nil {
				statuses <- 0
				return
			}
			resp.Body.Close()
			statuses <- resp.StatusCode
		}()
	}
	wg.Wait()
	close(statuses)
	got := []int{}
	for status := range statuses {
		got = append(got, status)
	}
	slices.Sort(got)
	require.Equal(t, []int{http.StatusOK, http.StatusNotFound}, got)
	after, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)
	require.Equal(t, before+1, after)
}
