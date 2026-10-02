//go:build sqlite

package main

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/arran4/goa4web/core"
	corecommon "github.com/arran4/goa4web/core/common"
	privateforumhandlers "github.com/arran4/goa4web/handlers/privateforum"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/stretchr/testify/require"
)

func recordTaskEvents(bus *eventbus.Bus) <-chan eventbus.TaskEvent {
	events := make(chan eventbus.TaskEvent, 4)
	bus.SyncPublish = func(message eventbus.Message) {
		if evt, ok := message.(eventbus.TaskEvent); ok {
			events <- evt
		}
	}
	return events
}

func requireSingleTaskEvent(t *testing.T, events <-chan eventbus.TaskEvent) eventbus.TaskEvent {
	t.Helper()
	var evt eventbus.TaskEvent
	select {
	case evt = <-events:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for task event")
	}
	select {
	case extra := <-events:
		t.Fatalf("unexpected additional task event: %#v", extra)
	case <-time.After(50 * time.Millisecond):
	}
	return evt
}

func authenticatedSession(t *testing.T, serverURL string, client *http.Client) (*url.URL, *http.Request, map[any]any) {
	t.Helper()
	baseURL, err := url.Parse(serverURL)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, serverURL+"/", nil)
	for _, cookie := range client.Jar.Cookies(baseURL) {
		req.AddCookie(cookie)
	}
	session, err := core.Store.Get(req, core.SessionName)
	require.NoError(t, err)
	require.NotZero(t, session.Values["UID"])
	require.NotEmpty(t, session.Values["SessionRef"])
	return baseURL, req, session.Values
}

func expireAuthenticatedSession(t *testing.T, serverURL string, client *http.Client) {
	t.Helper()
	baseURL, req, _ := authenticatedSession(t, serverURL, client)
	session, err := core.Store.Get(req, core.SessionName)
	require.NoError(t, err)
	session.Values["ExpiryTime"] = time.Now().Add(-time.Hour).Unix()
	recorder := httptest.NewRecorder()
	require.NoError(t, session.Save(req, recorder))
	client.Jar.SetCookies(baseURL, recorder.Result().Cookies())
}

func revokeAuthenticatedSession(t *testing.T, serverURL string, client *http.Client, srvQueries interface {
	SystemDeleteSessionByID(context.Context, string) error
}) {
	t.Helper()
	_, _, values := authenticatedSession(t, serverURL, client)
	ref, ok := values["SessionRef"].(string)
	require.True(t, ok)
	require.NotEmpty(t, ref)
	require.NoError(t, srvQueries.SystemDeleteSessionByID(context.Background(), core.HashSessionRef(ref)))
}

func captureInterruptedPrivateTopic(t *testing.T, serverURL string, client *http.Client, form resumableForm, title string) string {
	t.Helper()
	values := privateTopicValues(form, title, "bob")
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	resp := postEncoded(t, client, serverURL+"/private/topic/new", values)
	require.Equal(t, http.StatusSeeOther, resp.StatusCode)
	requireNoStore(t, resp)
	require.Equal(t, "/login", mustParseURL(t, resp.Header.Get("Location")).Path)
	token := extractResumeToken(resp.Header.Get("Location"))
	require.NotEmpty(t, token)
	resp.Body.Close()
	return token
}

func resumeInterruptedPrivateTopic(t *testing.T, serverURL string, client *http.Client, token string) {
	t.Helper()
	client.CheckRedirect = nil
	loginUserFunc(t, serverURL, "alice", "alice-test", client)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	csrfToken, interstitial := resumeCSRF(t, serverURL, token, client)
	requireNoStore(t, interstitial)
	resume := postEncoded(t, client, serverURL+"/resume", url.Values{
		"token": {token}, "operation": {"resume"}, "gorilla.csrf.Token": {csrfToken},
	})
	require.Equal(t, http.StatusOK, resume.StatusCode)
	requireNoStore(t, resume)
	resume.Body.Close()
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(raw)
	require.NoError(t, err)
	return parsed
}

func TestResume_GenuineExpiredSessionCapturedAndResumed(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	client := createClient()
	loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
	form := fetchResumableForm(t, httpServer.URL, client)
	before, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)

	expireAuthenticatedSession(t, httpServer.URL, client)
	token := captureInterruptedPrivateTopic(t, httpServer.URL, client, form, "expired session resume")
	resumeInterruptedPrivateTopic(t, httpServer.URL, client, token)

	after, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)
	require.Equal(t, before+1, after)
}

func TestResume_AuthoritativeSessionRevocationCapturedAndResumed(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	client := createClient()
	loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
	form := fetchResumableForm(t, httpServer.URL, client)
	before, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)

	revokeAuthenticatedSession(t, httpServer.URL, client, srv.Queries)
	token := captureInterruptedPrivateTopic(t, httpServer.URL, client, form, "revoked session resume")
	resumeInterruptedPrivateTopic(t, httpServer.URL, client, token)

	after, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)
	require.Equal(t, before+1, after)
}

func TestResume_InvalidatedSessionNonOptedInPostUsesNormalLoginRedirect(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	client := createClient()
	loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
	expireAuthenticatedSession(t, httpServer.URL, client)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }

	resp := postEncoded(t, client, httpServer.URL+"/private", url.Values{
		"task": {string(privateforumhandlers.TaskPrivateTopicCreate)}, "title": {"not opted in"},
	})
	require.Equal(t, http.StatusSeeOther, resp.StatusCode)
	requireNoStore(t, resp)
	location := mustParseURL(t, resp.Header.Get("Location"))
	require.Equal(t, "/login", location.Path)
	require.Equal(t, "/private", location.Query().Get("back"))
	require.Empty(t, extractResumeToken(resp.Header.Get("Location")))
	resp.Body.Close()
	var count int
	require.NoError(t, srv.DB.QueryRow("SELECT COUNT(*) FROM pending_actions").Scan(&count))
	require.Zero(t, count)
}

func TestResume_SuccessRestoresOriginalTaskEvent(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	client := createClient()
	loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
	token, _, captureResp := capturePrivateTopic(t, httpServer.URL, client, "resumed event identity", "?source=event-test")
	captureResp.Body.Close()
	client.CheckRedirect = nil
	loginUserFunc(t, httpServer.URL, "alice", "alice-test", client)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	csrfToken, interstitial := resumeCSRF(t, httpServer.URL, token, client)
	interstitial.Body.Close()

	events := recordTaskEvents(srv.Bus)
	resume := postEncoded(t, client, httpServer.URL+"/resume", url.Values{
		"token": {token}, "operation": {"resume"}, "gorilla.csrf.Token": {csrfToken},
	})
	require.Equal(t, http.StatusOK, resume.StatusCode)
	resume.Body.Close()

	evt := requireSingleTaskEvent(t, events)
	task, ok := evt.Task.(*privateforumhandlers.PrivateTopicCreateTask)
	require.True(t, ok, "event task type = %T", evt.Task)
	require.Equal(t, string(privateforumhandlers.TaskPrivateTopicCreate), task.Name())
	require.Equal(t, "/private/topic/new", evt.Path)
	require.Equal(t, eventbus.TaskOutcomeSuccess, evt.Outcome)
	var aliceID int32
	require.NoError(t, srv.DB.QueryRow("SELECT idusers FROM users WHERE username = ?", "alice").Scan(&aliceID))
	require.Equal(t, aliceID, evt.UserID)

	provider, ok := evt.Task.(notif.AutoSubscribeProvider)
	require.True(t, ok, "resumed task must retain AutoSubscribeProvider behavior")
	actionName, subscriptionPath, err := provider.AutoSubscribePath(evt)
	require.NoError(t, err)
	require.Equal(t, string(privateforumhandlers.TaskPrivateTopicCreate), actionName)
	require.Equal(t, "/private/topic/new", subscriptionPath)

	var topicID int32
	var title string
	require.NoError(t, srv.DB.QueryRow("SELECT idforumtopic, title FROM forumtopic WHERE title = ?", "resumed event identity").Scan(&topicID, &title))
	require.Equal(t, "resumed event identity", title)
	var subscriptionCount int
	require.NoError(t, srv.DB.QueryRow(
		"SELECT COUNT(*) FROM subscriptions WHERE pattern = ? AND method = 'internal'",
		corecommon.TopicSubscriptionPattern(topicID, true),
	).Scan(&subscriptionCount))
	require.Equal(t, 2, subscriptionCount, "creator and participant subscriptions must match normal topic creation")
	var consumed sql.NullTime
	require.NoError(t, srv.DB.QueryRow("SELECT consumed_at FROM pending_actions WHERE id = ?", hashOpaque(token)).Scan(&consumed))
	require.True(t, consumed.Valid)
}
