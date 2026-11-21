//go:build sqlite || sqlite3

package main

import (
	"context"
	"html"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email/mock"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/scenario"
	"github.com/arran4/goa4web/workers/emailqueue"
	"github.com/stretchr/testify/require"
)

func TestScenarioGroupsUnreadRepliesAndRechecksAccess(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := context.Background()
	actor := func(name string) int32 {
		id, ok := env.registry.ResolveUser(name)
		require.True(t, ok)
		return id
	}
	alice, bob, carol := actor("alice"), actor("bob"), actor("carol")
	staff := requireScenarioID(t, env.registry, scenario.RefTypeForum, "staff-room")
	thread := requireScenarioID(t, env.registry, scenario.RefTypeThread, "staff-welcome")
	otherTopic := requireScenarioID(t, env.registry, scenario.RefTypeForum, "coordination")
	otherThread := requireScenarioID(t, env.registry, scenario.RefTypeThread, "coordination-plan")
	aliceCD := env.cd.ForUser(alice)
	readHistory := func(topicID, threadID int32) {
		err := aliceCD.SubscribeForum(ctx, common.SubscribeForumParams{ActorID: alice, TopicID: topicID, ThreadID: threadID})
		require.NoError(t, err)
		comments, err := aliceCD.ForUser(alice).ThreadComments(threadID)
		require.NoError(t, err)
		require.NotEmpty(t, comments)
		last := comments[len(comments)-1]
		mustRead := common.ReadForumThreadParams{ActorID: alice, ThreadID: threadID, LastActivityID: aliceCD.CommentDisplayActivity(last)}
		require.NoError(t, aliceCD.ReadForumThread(ctx, mustRead))
	}
	readHistory(staff, thread)
	readHistory(otherTopic, otherThread)

	reply := func(actorID, topicID, threadID int32, text string) *common.ReplyForumThreadResult {
		result, err := env.cd.ForUser(actorID).ReplyForumThread(ctx, common.ReplyForumThreadParams{
			ActorID: actorID, TopicID: topicID, ThreadID: threadID, LanguageID: 1, Text: text, At: time.Now().UTC(),
			Private: true, EnforceHandler: true, SynchronousSideEffects: true,
		})
		require.NoError(t, err)
		return result
	}
	first := reply(bob, staff, thread, "First new staff reply")
	second := reply(bob, staff, thread, "Second staff reply is an append")
	require.True(t, second.Appended)
	require.Equal(t, first.CommentID, second.CommentID)
	require.NoError(t, env.cd.ForumReplyNotifier().NotifyForumReply(ctx, second.Intent), "retry the same activity")
	reply(carol, otherTopic, otherThread, "An unrelated reply")

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := &http.Client{Jar: jar}
	doScenarioLogin(t, env.http, client, "alice", "alice-test")
	inbox := env.http.URL + "/usr/notifications"
	fragment := inbox + "/fragment/list"
	body := scenarioHTTPGet(t, client, fragment)
	require.Contains(t, body, "2 new replies in Staff Room")
	require.Contains(t, body, "1 new reply in Coordination")
	require.NotContains(t, body, "3 new replies in Staff Room")
	for _, suffix := range []string{"/rss", "/atom"} {
		feed := scenarioHTTPGet(t, client, inbox+suffix)
		require.Contains(t, feed, "2 new replies in Staff Room")
	}

	notifier := notifications.New(notifications.WithQueries(env.cd.Queries()), notifications.WithConfig(env.server.Config))
	require.NoError(t, notifier.SendDigestToUser(ctx, alice, "alice@example.test", true, false, notifications.DigestDaily))
	queued, err := env.cd.Queries().SystemListPendingEmails(ctx, db.SystemListPendingEmailsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, queued, 1)
	require.True(t, queued[0].DigestDeliveryID.Valid)
	require.Contains(t, queued[0].Body, "2 new replies in Staff Room")
	require.Contains(t, queued[0].Body, "1 new reply in Coordination")
	require.NotContains(t, queued[0].Body, "First new staff reply")
	memberCount, err := env.cd.Queries().SystemCountDigestDeliveryMembers(ctx, queued[0].DigestDeliveryID.Int32)
	require.NoError(t, err)
	require.Equal(t, int64(3), memberCount)

	snapshot := regexp.MustCompile(`name="group" value="([^"]+)"[^>]*>\s*[^<]*<a[^>]+>2 new replies in Staff Room`).FindStringSubmatch(body)
	require.Len(t, snapshot, 2, "staff reply group should have a dismissal snapshot")
	marker, err := aliceCD.ForUser(alice).ThreadReadMarkerPosition(thread)
	require.NoError(t, err)
	third := reply(bob, staff, thread, "Reply written after Alice's inbox snapshot")
	require.NotEqual(t, second.ActivityID, third.ActivityID)
	inboxPage := scenarioHTTPGet(t, client, inbox)
	scenarioHTTPPostForm(t, client, inbox+"/dismiss", url.Values{
		"task": {"Dismiss"}, "group": {html.UnescapeString(snapshot[1])},
		"gorilla.csrf.Token": {scenarioCSRFToken(t, inboxPage)},
	})
	body = scenarioHTTPGet(t, client, fragment)
	require.NotContains(t, body, "2 new replies in Staff Room")
	require.Contains(t, body, "1 new reply in Staff Room")
	markerAfter, err := aliceCD.ForUser(alice).ThreadReadMarkerPosition(thread)
	require.NoError(t, err)
	require.Equal(t, marker, markerAfter, "dismissing the group leaves the thread read position untouched")

	var grantID int32
	require.NoError(t, env.db.QueryRowContext(ctx, "SELECT id FROM grants WHERE section = 'privateforum_thread' AND action = 'view' AND item_id = ? AND user_id = ?", thread, alice).Scan(&grantID))
	require.NoError(t, aliceCD.RevokeForumThread(grantID))
	for _, target := range []string{fragment, inbox + "/rss", inbox + "/atom", fragment + "?all=1"} {
		body := scenarioHTTPGet(t, client, target)
		require.False(t, strings.Contains(body, "Staff Room"), "revoked thread appears at %s", target)
	}
	// A queued digest must be suppressed before contacting the provider when
	// any membership is revoked. Suppression must leave notification state intact.
	cfg := *env.server.Config
	cfg.EmailEnabled = true
	provider := &mock.Provider{}
	require.True(t, emailqueue.ProcessPendingEmail(ctx, env.cd.Queries(), provider, nil, &cfg))
	require.Empty(t, provider.Messages)
	pending, err := env.cd.Queries().SystemListPendingEmails(ctx, db.SystemListPendingEmailsParams{Limit: 10})
	require.NoError(t, err)
	require.Empty(t, pending)
	body = scenarioHTTPGet(t, client, fragment)
	require.Contains(t, body, "1 new reply in Coordination")
}
