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
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/testdata/scenarios"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestScenarioPrivateForumLiveReplies(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	serve, closeRoot := newTestScenarioServeCmd(t, []string{"goa4web", "scenario", "serve", "100-private-forum"})
	defer closeRoot()
	serve.fsys = scenarios.FS
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
		return client
	}
	alice, bob, carol := clientFor("alice"), clientFor("bob"), clientFor("carol")
	topicID, threadID := getTopicAndThreadIDForUser(ctx, t, sqlDB, "alice")
	staffPath := fmt.Sprintf("/private/topic/%d/thread/%d", topicID, threadID)
	page := scenarioHTTPGet(t, alice, ts.URL+staffPath)
	target := regexp.MustCompile(`data-live-url="([^"]+/live\?after=[0-9]+)"`).FindStringSubmatch(page)
	require.Len(t, target, 2, "thread page must advertise its rendered activity position")
	fragmentURL := ts.URL + target[1]

	for _, tc := range []struct {
		client *http.Client
		status int
	}{{alice, http.StatusNoContent}, {carol, http.StatusNotFound}} {
		resp, err := tc.client.Get(fragmentURL)
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, tc.status, resp.StatusCode)
	}

	dial := func(client *http.Client, requestedThread int32) *websocket.Conn {
		d := *websocket.DefaultDialer
		d.Jar = client.Jar
		endpoint := fmt.Sprintf("ws%s/ws/events?thread=%d", strings.TrimPrefix(ts.URL, "http"), requestedThread)
		socket, resp, err := d.DialContext(ctx, endpoint, nil)
		if resp != nil {
			resp.Body.Close()
		}
		require.NoError(t, err)
		t.Cleanup(func() { socket.Close() })
		return socket
	}
	aliceSocket := dial(alice, threadID)
	bobSocket := dial(bob, threadID)
	carolSocket := dial(carol, threadID)
	wrongThreadSocket := dial(alice, threadID+1)

	postReply := func(author *http.Client, text string) {
		bobPage := scenarioHTTPGet(t, author, ts.URL+staffPath)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+staffPath+"/reply", strings.NewReader(url.Values{
			"replytext": {text}, "task": {"Reply"}, "gorilla.csrf.Token": {scenarioCSRFToken(t, bobPage)},
		}.Encode()))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Referer", ts.URL+staffPath)
		client := *author
		client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, http.StatusSeeOther, resp.StatusCode)
	}
	readSignal := func(socket *websocket.Conn, kind string) {
		require.NoError(t, socket.SetReadDeadline(time.Now().Add(2*time.Second)))
		_, raw, err := socket.ReadMessage()
		require.NoError(t, err)
		require.Contains(t, string(raw), `"kind":"`+kind+`"`)
		for _, field := range []string{"thread_id", "comment_id", "data", staffPath, "private live reply"} {
			require.NotContains(t, string(raw), field)
		}
	}

	q := db.NewForDriver(sqlDB, "sqlite3")
	aliceUser, err := q.SystemGetUserByUsername(ctx, sql.NullString{String: "alice", Valid: true})
	require.NoError(t, err)
	aliceCD := common.NewCoreData(ctx, q, srv.Config, common.WithUserID(aliceUser.Idusers))
	before, err := aliceCD.ThreadReadMarker(threadID)
	require.NoError(t, err)

	postReply(bob, "private live reply <script>alert(1)</script>")
	readSignal(aliceSocket, "thread-reply")
	readSignal(bobSocket, "thread-reply")
	resp, err := alice.Get(fragmentURL)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, string(body), "New replies are available")
	require.Contains(t, string(body), staffPath)
	require.NotContains(t, string(body), "private live reply")
	require.Contains(t, resp.Header.Get("Cache-Control"), "no-store")
	after, err := aliceCD.ForUser(aliceUser.Idusers).ThreadReadMarker(threadID)
	require.NoError(t, err)
	require.Equal(t, before, after)

	// Use recipient-only inbox signals as ordered barriers. A reply processed on
	// the wrong scope or delivered to a nonparticipant would be seen first.
	carolUser, err := q.SystemGetUserByUsername(ctx, sql.NullString{String: "carol", Valid: true})
	require.NoError(t, err)
	barrier := func(socket *websocket.Conn, user int32) {
		require.NoError(t, srv.Bus.Publish(eventbus.NotificationsChanged{RecipientID: user, Time: time.Now()}))
		readSignal(socket, "notification")
	}
	barrier(carolSocket, carolUser.Idusers)
	barrier(wrongThreadSocket, aliceUser.Idusers)

	var grantID int32
	require.NoError(t, sqlDB.QueryRowContext(ctx, "SELECT id FROM grants WHERE section = 'privateforum_thread' AND action = 'view' AND item_id = ? AND user_id = ?", threadID, aliceUser.Idusers).Scan(&grantID))
	require.NoError(t, aliceCD.RevokeForumThread(grantID))
	postReply(bob, "second private live reply")
	barrier(aliceSocket, aliceUser.Idusers)
	resp, err = alice.Get(fragmentURL)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	bobUser, err := q.SystemGetUserByUsername(ctx, sql.NullString{String: "bob", Valid: true})
	require.NoError(t, err)
	readSignal(bobSocket, "thread-reply") // Bob still received his second reply.
	require.NoError(t, q.AdminScrubUser(ctx, db.AdminScrubUserParams{
		Idusers: bobUser.Idusers, Username: sql.NullString{String: "deactivated-bob", Valid: true},
	}))
	_, err = aliceCD.GrantForumThread(threadID, sql.NullInt32{Int32: aliceUser.Idusers, Valid: true}, sql.NullInt32{}, "view")
	require.NoError(t, err)
	postReply(alice, "private live reply after deactivation")
	barrier(bobSocket, bobUser.Idusers)
}
