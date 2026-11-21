//go:build sqlite || sqlite3

package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers/forum"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/stretchr/testify/require"
)

func TestScenarioPublicForumFollowOnlyChangesSelectedScope(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()
	alice, ok := env.registry.ResolveUser("alice")
	require.True(t, ok)
	forumURL := env.http.URL + "/forum"
	require.NotContains(t, scenarioHTTPGet(t, env.http.Client(), forumURL), "Follow new threads")
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := &http.Client{Jar: jar}
	doScenarioLogin(t, env.http, client, "alice", "alice-test")
	body := scenarioHTTPGet(t, client, forumURL)
	require.Contains(t, body, "Follow new threads")
	before, err := env.cd.Queries().ListSubscriptionsByUser(ctx, alice)
	require.NoError(t, err)
	badForm := url.Values{"follow": {"yes"}}
	response, err := client.PostForm(forumURL+"/follow", badForm)
	require.NoError(t, err)
	response.Body.Close()
	require.Equal(t, http.StatusForbidden, response.StatusCode)
	post := func(follow string) {
		form := url.Values{"follow": {follow}, "gorilla.csrf.Token": {scenarioCSRFToken(t, body)}}
		scenarioHTTPPostForm(t, client, forumURL+"/follow", form)
		body = scenarioHTTPGet(t, client, forumURL)
	}
	post("yes")
	post("yes")
	require.Contains(t, body, "Unfollow new threads")
	withFollow, err := env.cd.Queries().ListSubscriptionsByUser(ctx, alice)
	require.NoError(t, err)
	require.Len(t, withFollow, len(before)+1)

	// Restrict the two public topics independently so only one can send Alice a notice.
	_, err = env.db.ExecContext(ctx, "DELETE FROM grants WHERE section = 'forum' AND action = 'view' AND (item = 'topic' OR item IS NULL)")
	require.NoError(t, err)
	queries := env.cd.Queries()
	bob, ok := env.registry.ResolveUser("bob")
	require.True(t, ok)
	addTopic := func(title string, aliceCanView bool) int32 {
		id, err := queries.AdminCreateForumTopic(ctx, db.AdminCreateForumTopicParams{
			LanguageID: sql.NullInt32{Int32: 1, Valid: true},
			Title:      sql.NullString{String: title, Valid: true}, Description: sql.NullString{String: "Test topic", Valid: true},
		})
		require.NoError(t, err)
		for _, action := range []string{"view", "post"} {
			_, err := env.cd.GrantForumTopic(int32(id), sql.NullInt32{Int32: bob, Valid: true}, sql.NullInt32{}, action)
			require.NoError(t, err)
		}
		if aliceCanView {
			_, err := env.cd.GrantForumTopic(int32(id), sql.NullInt32{Int32: alice, Valid: true}, sql.NullInt32{}, "view")
			require.NoError(t, err)
		}
		return int32(id)
	}
	allowed := addTopic("Accessible announcements", true)
	protected := addTopic("Restricted discussion", false)
	send := func(topic int32, text string) string {
		evt := &eventbus.TaskEvent{UserID: bob, Task: forum.CreateThreadTaskHandler, Outcome: eventbus.TaskOutcomeSuccess}
		actor := common.NewCoreData(ctx, queries, env.server.Config, common.WithUserID(bob), common.WithEvent(evt))
		result, err := actor.CreateForumThread(ctx, common.CreateForumThreadParams{
			ActorID: bob, TopicID: topic, Text: text, LanguageID: 1, EnforceHandler: true, SynchronousSideEffects: true,
		})
		require.NoError(t, err)
		n := notifications.New(notifications.WithQueries(queries), notifications.WithConfig(env.server.Config))
		require.NoError(t, n.ProcessEvent(ctx, *evt, nil))
		return fmt.Sprintf("create thread:comment:%d", result.CommentID)
	}
	visibleKey := send(allowed, "A new conversation Alice can read")
	hiddenKey := send(protected, "A new conversation Alice cannot read")
	_, err = env.cd.ForUser(alice).NotificationIDByActivity(ctx, visibleKey)
	require.NoError(t, err)
	_, err = env.cd.ForUser(alice).NotificationIDByActivity(ctx, hiddenKey)
	require.ErrorIs(t, err, sql.ErrNoRows)
	inbox := scenarioHTTPGet(t, client, env.http.URL+"/usr/notifications")
	require.Contains(t, inbox, "Accessible announcements")
	require.NotContains(t, inbox, "Restricted discussion")
	allowedURL := fmt.Sprintf("%s/topic/%d", forumURL, allowed)
	allowedPage := scenarioHTTPGet(t, client, allowedURL)
	require.Contains(t, allowedPage, "You follow new threads from the")
	scenarioHTTPPostForm(t, client, allowedURL+"/subscribe", url.Values{
		"task": {"Subscribe To Topic"}, "gorilla.csrf.Token": {scenarioCSRFToken(t, allowedPage)},
	})
	post("no")
	require.Contains(t, body, "Follow new threads")
	allowedPage = scenarioHTTPGet(t, client, allowedURL)
	require.NotContains(t, allowedPage, "You follow new threads from the")
	require.Contains(t, allowedPage, "Remove direct topic follow")
	scenarioHTTPPostForm(t, client, allowedURL+"/unsubscribe", url.Values{
		"task": {"Unsubscribe From Topic"}, "gorilla.csrf.Token": {scenarioCSRFToken(t, allowedPage)},
	})
	after, err := env.cd.Queries().ListSubscriptionsByUser(ctx, alice)
	require.NoError(t, err)
	require.Equal(t, before, after, "removing the forum and topic follows preserves other subscriptions")
}
