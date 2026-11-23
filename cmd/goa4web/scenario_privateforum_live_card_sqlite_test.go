//go:build sqlite || sqlite3

package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/externallink"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/opengraph"
	"github.com/arran4/goa4web/testdata/scenarios"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestScenarioPrivateForumLiveExternalCard(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	serve, closeRoot := newTestScenarioServeCmd(t, []string{"goa4web", "scenario", "serve", "100-private-forum"})
	defer closeRoot()
	serve.fsys = scenarios.FS
	// A temporary file is required to share durable work across independent DB handles.
	dsn := filepath.Join(t.TempDir(), "live-cards.sqlite")
	webDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)
	defer webDB.Close()
	serve.dbConn = webDB
	srv, sqlDB, cleanup, err := serve.Bootstrap(ctx)
	require.NoError(t, err)
	defer cleanup()
	ts := httptest.NewServer(srv.Router)
	defer ts.Close()

	clientFor := func(username string) *http.Client {
		jar, err := cookiejar.New(nil)
		require.NoError(t, err)
		client := &http.Client{Jar: jar}
		doScenarioLogin(t, ts, client, username, username+"-test")
		client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
		return client
	}
	alice, carol := clientFor("alice"), clientFor("carol")
	topicID, threadID := getTopicAndThreadIDForUser(ctx, t, sqlDB, "alice")
	threadURL := ts.URL + fmt.Sprintf("/private/topic/%d/thread/%d", topicID, threadID)
	page := scenarioHTTPGet(t, alice, threadURL)

	rawURL := "https://preview.example.test/new-card"
	form := url.Values{
		"replytext": {"[link " + rawURL + "]"}, "task": {"Reply"}, "gorilla.csrf.Token": {scenarioCSRFToken(t, page)},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, threadURL+"/reply", strings.NewReader(form.Encode()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", threadURL)
	resp, err := alice.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusSeeOther, resp.StatusCode)

	page = scenarioHTTPGet(t, alice, threadURL)
	require.True(t, strings.Contains(page, "Fetching link preview"), "thread must show the pending card")
	target := regexp.MustCompile(`data-live-card="([0-9]+-[a-f0-9]+)" data-live-url="([^"]+)"`).FindStringSubmatch(page)
	require.Len(t, target, 3)
	fragmentURL := ts.URL + target[2]

	fragment := func(client *http.Client, url string, status int) string {
		resp, err := client.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, status, resp.StatusCode)
		return string(body)
	}
	require.Contains(t, fragment(alice, fragmentURL, http.StatusOK), "Fetching link preview")
	require.NotContains(t, fragment(carol, fragmentURL, http.StatusNotFound), rawURL)
	commentKey, _, _ := strings.Cut(target[1], "-")
	forged := strings.Replace(fragmentURL, target[1], commentKey+"-"+strings.Repeat("f", 64), 1)
	require.NotContains(t, fragment(alice, forged, http.StatusNotFound), rawURL)

	dial := func(client *http.Client, key string) *websocket.Conn {
		dialer := *websocket.DefaultDialer
		dialer.Jar = client.Jar
		wsURL := fmt.Sprintf("ws%s/ws/events?thread=%d&card=%s", strings.TrimPrefix(ts.URL, "http"), threadID, key)
		socket, _, err := dialer.DialContext(ctx, wsURL, nil)
		require.NoError(t, err)
		t.Cleanup(func() { socket.Close() })
		return socket
	}
	unrelatedSocket := dial(carol, target[1])
	forgedSocket := dial(alice, commentKey+"-"+strings.Repeat("f", 64))
	socket := dial(alice, target[1])
	defer socket.Close()

	// A separate worker service uses its own DB handle and has no browser bus.
	workerDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)
	defer workerDB.Close()
	q := db.NewForDriver(workerDB, "sqlite3")
	worker := jobs.NewService(q)
	claim, err := worker.ClaimNext(ctx, "standalone-test", jobs.JobTypeExternalLink)
	require.NoError(t, err)
	require.NotNil(t, claim)
	require.Equal(t, rawURL, claim.Job.TargetKey)
	const title = "<img src=x onerror=bad()>"
	_, err = externallink.RefreshMetadata(ctx, q, rawURL, func(ctx context.Context) (int32, error) {
		return externallink.ResolveRecord(ctx, q, rawURL)
	}, func(context.Context, string) (*opengraph.Info, error) {
		return &opengraph.Info{Title: title, Description: "Completed by a separate worker"}, nil
	}, nil)
	require.NoError(t, err)
	completed, err := worker.CompleteFenced(ctx, claim)
	require.NoError(t, err)
	require.True(t, completed)

	require.NoError(t, socket.SetReadDeadline(time.Now().Add(7*time.Second)))
	_, payload, err := socket.ReadMessage()
	require.NoError(t, err)
	require.Contains(t, string(payload), `"kind":"external-link-card"`)
	require.NotContains(t, string(payload), rawURL)
	require.NotContains(t, string(payload), "standalone-test")
	for _, denied := range []*websocket.Conn{unrelatedSocket, forgedSocket} {
		require.NoError(t, denied.SetReadDeadline(time.Now().Add(300*time.Millisecond)))
		_, payload, err := denied.ReadMessage()
		require.Error(t, err, "denied watcher received %s", payload)
	}
	body := fragment(alice, fragmentURL, http.StatusOK)
	require.Contains(t, body, "Completed by a separate worker")
	require.Contains(t, body, "&lt;img src=x onerror=bad()&gt;")
	require.NotContains(t, body, title)
	require.NotContains(t, body, "Fetching link preview")

	viewer := common.NewCoreData(ctx, q, srv.Config)
	var grantID int32
	require.NoError(t, sqlDB.QueryRowContext(ctx, "SELECT g.id FROM grants g JOIN users u ON u.idusers = g.user_id WHERE g.section = 'privateforum_thread' AND g.action = 'view' AND g.item_id = ? AND u.username = 'alice'", threadID).Scan(&grantID))
	require.NoError(t, viewer.RevokeForumThread(grantID))
	require.NotContains(t, fragment(alice, fragmentURL, http.StatusNotFound), rawURL)
}
