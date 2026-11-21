//go:build sqlite || sqlite3

package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/scenario"
	"github.com/stretchr/testify/require"
)

func TestScenarioSubscriptionOverviewUsesCurrentPrivateAccess(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()
	alice, ok := env.registry.ResolveUser("alice")
	require.True(t, ok)
	topic := requireScenarioID(t, env.registry, scenario.RefTypeForum, "staff-room")
	thread := requireScenarioID(t, env.registry, scenario.RefTypeThread, "staff-welcome")
	aliceCD := env.cd.ForUser(alice)
	q := env.cd.Queries()
	require.NoError(t, aliceCD.SetPublicForumNewThreadFollow(true))
	pattern := common.ThreadSubscriptionPattern(topic, thread, true)
	require.NoError(t, common.EnsureSubscriptionIdempotent(ctx, q, alice, pattern, "internal"))
	require.NoError(t, common.EnsureSubscriptionIdempotent(ctx, q, alice, pattern, "email"))
	require.NoError(t, aliceCD.SaveEmail(alice, false, false))
	require.NoError(t, common.EnsureSubscriptionIdempotent(ctx, q, alice, strings.TrimSuffix(pattern, "/*"), "internal"))
	require.NoError(t, aliceCD.SubscribeTopic(alice, topic, true))
	require.NoError(t, common.EnsureSubscriptionIdempotent(ctx, q, alice, "post:/blog/*", "internal"))
	bob, ok := env.registry.ResolveUser("bob")
	require.True(t, ok)
	var outsideOverview int32
	for range 101 {
		result, err := env.cd.ForUser(bob).CreateForumThread(ctx, common.CreateForumThreadParams{
			ActorID: bob, TopicID: topic, Text: "Another thread under the followed topic", LanguageID: 1,
			Private: true, EnforceHandler: true, SynchronousSideEffects: true,
		})
		require.NoError(t, err)
		outsideOverview = result.ThreadID
	}

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := &http.Client{Jar: jar}
	doScenarioLogin(t, env.http, client, "alice", "alice-test")
	overview := env.http.URL + "/usr/subscriptions"
	body := scenarioHTTPGet(t, client, overview)
	threadLink := fmt.Sprintf(`href="/private/topic/%d/thread/%d"`, topic, thread)
	require.Contains(t, body, "Your forum follows")
	require.Contains(t, body, "New Blog Post")
	require.Contains(t, body, "Public forum")
	require.Equal(t, 1, strings.Count(body, threadLink), "legacy aliases represent one thread scope")
	require.NotContains(t, body, fmt.Sprintf(`href="/private/topic/%d/thread/%d"`, topic, outsideOverview))
	require.NotContains(t, body, "Another thread under the followed topic")
	require.Contains(t, body, "Automatic follow when you post is off")
	require.Contains(t, body, "Email replies paused")
	require.NotContains(t, body, "Global Controls")
	require.Contains(t, scenarioHTTPGet(t, client, overview+"?mode=advanced"), "Global Controls")

	var grantID int32
	require.NoError(t, env.db.QueryRowContext(ctx, "SELECT id FROM grants WHERE section = 'privateforum_thread' AND action = 'view' AND item_id = ? AND user_id = ?", thread, alice).Scan(&grantID))
	require.NoError(t, aliceCD.RevokeForumThread(grantID))
	body = scenarioHTTPGet(t, client, overview)
	require.NotContains(t, body, threadLink)
	require.NotContains(t, body, "Welcome to the staff room")
	require.Contains(t, body, "Public forum")
}
