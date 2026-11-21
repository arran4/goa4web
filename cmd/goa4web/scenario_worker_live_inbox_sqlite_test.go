//go:build sqlite || sqlite3

package main

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/scenario"
	live "github.com/arran4/goa4web/internal/websocket"
	"github.com/arran4/goa4web/workers"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestScenarioLiveInboxObservesStandaloneReplyWorker(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()

	dialAs := func(user string) (*http.Client, *websocket.Conn) {
		t.Helper()
		jar, err := cookiejar.New(nil)
		require.NoError(t, err)
		client := &http.Client{Jar: jar}
		doScenarioLogin(t, env.http, client, user, user+"-test")
		dialer := *websocket.DefaultDialer
		dialer.Jar = jar
		socket, resp, err := dialer.DialContext(ctx, "ws"+strings.TrimPrefix(env.http.URL, "http")+"/ws/events", nil)
		if resp != nil {
			resp.Body.Close()
		}
		require.NoError(t, err)
		t.Cleanup(func() { _ = socket.Close() })
		return client, socket
	}
	_, authorSocket := dialAs("bob")
	reader, readerSocket := dialAs("alice")
	author, ok := env.registry.ResolveUser("bob")
	require.True(t, ok)
	recipient, ok := env.registry.ResolveUser("alice")
	require.True(t, ok)
	topic := requireScenarioID(t, env.registry, scenario.RefTypeForum, "staff-room")
	thread := requireScenarioID(t, env.registry, scenario.RefTypeThread, "staff-welcome")
	require.NoError(t, env.cd.SubscribeForum(ctx, common.SubscribeForumParams{ActorID: recipient, TopicID: topic, ThreadID: thread}))

	q := env.cd.Queries()
	app := common.NewCoreData(ctx, q, env.server.Config, common.WithJobService(jobs.NewService(q)))
	_, err := app.ForUser(author).ReplyForumThread(ctx, common.ReplyForumThreadParams{
		ActorID: author, TopicID: topic, ThreadID: thread, LanguageID: 1, Text: "private standalone reply",
		Private: true, EnforceHandler: true,
	})
	require.NoError(t, err)
	group := workers.StartDurableWorkers(ctx, q, env.server.Config, jobs.NewService(q), "without-web-bus")
	defer func() {
		group.Cancel()
		waitCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		require.NoError(t, group.Wait(waitCtx))
	}()

	require.NoError(t, readerSocket.SetReadDeadline(time.Now().Add(8*time.Second)))
	_, frame, err := readerSocket.ReadMessage()
	require.NoError(t, err, "an already connected browser should see the committed inbox change")
	require.Contains(t, string(frame), `"kind":"notification"`)
	require.NotContains(t, string(frame), "private standalone reply")
	require.NotContains(t, string(frame), "/private/")
	require.NotContains(t, string(frame), "recipient")
	inbox := scenarioHTTPGet(t, reader, env.http.URL+"/usr/notifications/fragment/list")
	require.Contains(t, inbox, "1 new reply in Staff Room")

	// The author's next signal must be their own bus barrier.
	barrier := time.Now().UTC()
	require.NoError(t, env.server.Bus.Publish(eventbus.NotificationsChanged{RecipientID: author, Time: barrier}))
	require.NoError(t, authorSocket.SetReadDeadline(time.Now().Add(2*time.Second)))
	var signal live.LiveEvent
	require.NoError(t, authorSocket.ReadJSON(&signal))
	require.Equal(t, live.LiveEventKindNotification, signal.Kind)
	require.True(t, signal.Time.Equal(barrier), "the author must not receive the follower's invalidation")
}
