//go:build sqlite || sqlite3

package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/scenario"
	"github.com/stretchr/testify/require"
)

func TestE2EPrivateForumAPIReadActivity(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := context.Background()

	createKey := func(user, name, scopes string) string {
		id, ok := env.registry.ResolveUser(user)
		require.True(t, ok)
		token := user + "-" + name + "-test"
		digest := sha256.Sum256([]byte(token))
		_, err := env.cd.Queries().CreateAPIKey(ctx, db.CreateAPIKeyParams{
			UsersIdusers: id,
			ApiKey:       hex.EncodeToString(digest[:]),
			Name:         name,
			Scopes:       scopes,
		})
		require.NoError(t, err)
		return token
	}
	const bothScopes = "private_forum:read,private_forum:write" // Read and post using the same test client.
	alice := createKey("alice", "read-activity", bothScopes)
	bob := createKey("bob", "reply-activity", bothScopes)
	readOnly := createKey("alice", "read-only", "private_forum:read")
	dave := createKey("dave", "nonparticipant", bothScopes)
	aliceID, _ := env.registry.ResolveUser("alice")
	staffTopic := requireScenarioID(t, env.registry, scenario.RefTypeForum, "staff-room")
	staffThread := requireScenarioID(t, env.registry, scenario.RefTypeThread, "staff-welcome")
	otherTopic := requireScenarioID(t, env.registry, scenario.RefTypeForum, "coordination")
	otherThread := requireScenarioID(t, env.registry, scenario.RefTypeThread, "coordination-plan")
	path := fmt.Sprintf("/private/api/topic/%d/thread/%d", staffTopic, staffThread)
	otherPath := fmt.Sprintf("/private/api/topic/%d/thread/%d", otherTopic, otherThread)
	require.NoError(t, env.cd.SubscribeForum(ctx, common.SubscribeForumParams{
		ActorID: aliceID, TopicID: staffTopic, ThreadID: staffThread,
	}))
	// Drain the HTTP-produced intents through the usual notifier without
	// relying on a background worker's timing or seeding notification rows.
	intents := env.server.Bus.SubscribeWithOptions(
		eventbus.WithTypes(eventbus.TaskMessageType),
		eventbus.WithBufferSize(8),
		eventbus.WithFilter(func(message eventbus.Message) bool {
			event, ok := message.(eventbus.TaskEvent)
			if !ok {
				return false
			}
			_, ok = event.Data[common.ForumReplyIntentEventKey].(common.ForumReplyNotificationIntent)
			return ok
		}),
	)
	defer intents.Close()

	unreadNotificationPositions := func(activities ...map[string]any) []int64 {
		rows, err := env.cd.Queries().ListUnreadNotificationsForLister(ctx, db.ListUnreadNotificationsForListerParams{
			ListerID: aliceID, Limit: 100,
		})
		require.NoError(t, err)
		positions := map[int64]bool{}
		for _, row := range rows {
			if row.ResourceType.String == "thread" && row.ResourceID.Int32 == staffThread && row.ActivityPosition.Valid {
				positions[row.ActivityPosition.Int64] = true
			}
		}
		var matched []int64
		for _, activity := range activities {
			pos, err := common.ParseActivityPosition(activity["activity_id"].(string))
			require.NoError(t, err)
			if positions[pos] {
				matched = append(matched, pos)
			}
		}
		return matched
	}
	marker := func() int64 {
		pos, err := env.cd.ForUser(aliceID).ThreadReadMarkerPosition(staffThread)
		require.NoError(t, err)
		return pos
	}
	unread := func() bool {
		rows, err := env.cd.ForUser(aliceID).UnreadPrivateThreads(100, 0, sql.NullInt32{Int32: staffTopic, Valid: true})
		require.NoError(t, err)
		for _, row := range rows {
			if row.Idforumthread == staffThread {
				return true
			}
		}
		return false
	}
	displayed := func(path string) map[string]any {
		resp, body := apiRequest(t, env, http.MethodGet, path, alice, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		comments := body["comments"].([]any)
		require.NotEmpty(t, comments)
		row := comments[len(comments)-1].(map[string]any)
		require.IsType(t, "", row["activity_id"])
		return row
	}
	post := func(text string) map[string]any {
		resp, body := apiRequest(t, env, http.MethodPost, path+"/reply", bob, url.Values{"text": {text}})
		require.Equal(t, http.StatusOK, resp.StatusCode, "%v", body)
		select {
		case envelope := <-intents.Channel():
			event := envelope.Msg.(eventbus.TaskEvent)
			intent := event.Data[common.ForumReplyIntentEventKey].(common.ForumReplyNotificationIntent)
			envelope.Ack()
			require.Equal(t, body["activity_id"], intent.ActivityID)
			require.NoError(t, env.cd.ForumReplyNotifier().NotifyForumReply(ctx, intent))
		case <-time.After(5 * time.Second):
			t.Fatal("REST reply did not publish its notification intent")
		}
		return body
	}
	mark := func(path, token, activity string, code int) {
		resp, body := apiRequest(t, env, http.MethodPost, path+"/read", token, url.Values{"last_activity": {activity}})
		require.Equal(t, code, resp.StatusCode, "%v", body)
	}

	before := marker()
	first := post("Reply before Alice fetches the thread")
	shown := displayed(path)
	require.Equal(t, first["activity_id"], shown["activity_id"])
	require.Equal(t, before, marker(), "GET must leave the read watermark untouched")
	second := post("Append after Alice fetched the thread")
	require.Equal(t, true, second["appended"])
	require.Equal(t, first["comment_id"], second["comment_id"])
	require.NotEqual(t, shown["activity_id"], second["activity_id"])
	third := post("Another append that Alice has yet to display")
	require.Equal(t, true, third["appended"])
	require.Equal(t, first["comment_id"], third["comment_id"])
	require.NotEqual(t, second["activity_id"], third["activity_id"])
	require.Len(t, unreadNotificationPositions(first, second, third), 3)

	mark(path, "invalid-test-key", shown["activity_id"].(string), http.StatusUnauthorized)
	missingKey, err := env.client.PostForm(env.http.URL+path+"/read", url.Values{"last_activity": {shown["activity_id"].(string)}})
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, missingKey.StatusCode, "anonymous form posts must still pass CSRF")
	missingKey.Body.Close()
	mark(path, readOnly, shown["activity_id"].(string), http.StatusForbidden)
	mark(path, dave, shown["activity_id"].(string), http.StatusNotFound)
	mark(fmt.Sprintf("/private/api/topic/%d/thread/%d", otherTopic, staffThread), alice, shown["activity_id"].(string), http.StatusNotFound)
	mark(path, alice, "", http.StatusBadRequest)
	mark(path, alice, "malformed", http.StatusBadRequest)
	mark(path, alice, fmt.Sprintf("comment:%d:append:-1", int(shown["id"].(float64))), http.StatusBadRequest)
	mark(path, alice, fmt.Sprintf("comment:%d:append:4294967297", int(shown["id"].(float64))), http.StatusBadRequest)
	mark(path, alice, displayed(otherPath)["activity_id"].(string), http.StatusNotFound)
	require.Equal(t, before, marker())

	old := shown["activity_id"].(string)
	mark(path, alice, old, http.StatusOK)
	oldPos, err := common.ParseActivityPosition(old)
	require.NoError(t, err)
	require.Equal(t, oldPos, marker(), "marking the displayed reply must leave both newer appends unread")
	require.True(t, unread(), "the thread should remain in the unread view until its displayed appends are acknowledged")
	require.Equal(t, unreadNotificationPositions(second, third), unreadNotificationPositions(first, second, third))
	require.Len(t, unreadNotificationPositions(second, third), 2)

	latest := displayed(path)["activity_id"].(string)
	require.Equal(t, third["activity_id"], latest)
	require.Equal(t, oldPos, marker())
	mark(path, alice, latest, http.StatusOK)
	latestPos, err := common.ParseActivityPosition(latest)
	require.NoError(t, err)
	require.Equal(t, latestPos, marker())
	require.False(t, unread(), "acknowledging the newest append removes the thread from the unread view")
	require.Empty(t, unreadNotificationPositions(first, second, third))
	mark(path, alice, old, http.StatusOK)
	require.Equal(t, latestPos, marker(), "an old page must not move the watermark backward")

	var grantID int32
	require.NoError(t, env.db.QueryRowContext(ctx, "SELECT id FROM grants WHERE section = 'privateforum_thread' AND action = 'view' AND item_id = ? AND user_id = ?", staffThread, aliceID).Scan(&grantID))
	require.NoError(t, env.cd.RevokeForumThread(grantID))
	mark(path, alice, latest, http.StatusNotFound)
	require.Equal(t, latestPos, marker())
}
