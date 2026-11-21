//go:build sqlite || sqlite3

package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/mail"
	"net/url"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/email/mock"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/scenario"
	"github.com/stretchr/testify/require"
)

func TestScenarioBatchesOptionalReplyEmailAtWindowBoundary(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()
	alice, found := env.registry.ResolveUser("alice")
	require.True(t, found)
	bob, found := env.registry.ResolveUser("bob")
	require.True(t, found)
	topic := requireScenarioID(t, env.registry, scenario.RefTypeForum, "staff-room")
	thread := requireScenarioID(t, env.registry, scenario.RefTypeThread, "staff-welcome")
	aliceCD := env.cd.ForUser(alice)
	require.NoError(t, aliceCD.SaveEmail(alice, true, true))
	require.NoError(t, aliceCD.AddUserEmail(alice, "alice-batch@example.test", "batch-test-code", time.Now().Add(time.Hour)))
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := &http.Client{Jar: jar}
	doScenarioLogin(t, env.http, client, "alice", "alice-test")
	verify := env.http.URL + "/usr/email/verify"
	form := scenarioHTTPGet(t, client, verify+"?code=batch-test-code")
	scenarioHTTPPostForm(t, client, verify, url.Values{
		"code": {"batch-test-code"}, "gorilla.csrf.Token": {scenarioCSRFToken(t, form)},
	})
	aliceUser, err := env.cd.Queries().SystemGetUserByID(ctx, alice)
	require.NoError(t, err)
	require.Equal(t, "alice-batch@example.test", aliceUser.Email.String)

	cfg := *env.server.Config
	cfg.EmailEnabled = true
	require.Equal(t, 15, cfg.ForumReplyEmailBatchMinutes)
	now := time.Date(2026, 11, 21, 10, 1, 0, 0, time.UTC)
	provider := &mock.Provider{}
	newWorker := func(q db.Querier, p email.Provider) *notifications.Notifier {
		return notifications.New(notifications.WithQueries(q), notifications.WithConfig(&cfg),
			notifications.WithReplyEmailClock(func() time.Time { return now }), notifications.WithEmailProvider(p))
	}
	queries := env.cd.Queries()
	producer := newWorker(queries, provider)
	core := common.NewCoreData(ctx, env.cd.Queries(), &cfg, common.WithForumReplyNotifier(producer))
	replyTo := func(actor, topic, thread int32, body string) *common.ReplyForumThreadResult {
		result, err := core.ForUser(actor).ReplyForumThread(ctx, common.ReplyForumThreadParams{
			ActorID: actor, TopicID: topic, ThreadID: thread, LanguageID: 1, Text: body,
			At: now, Private: true, EnforceHandler: true, SynchronousSideEffects: true,
		})
		require.NoError(t, err)
		return result
	}
	reply := func(actor int32, body string) *common.ReplyForumThreadResult {
		return replyTo(actor, topic, thread, body)
	}
	reply(alice, "Following the thread by participating")
	coordination := requireScenarioID(t, env.registry, scenario.RefTypeForum, "coordination")
	plan := requireScenarioID(t, env.registry, scenario.RefTypeThread, "coordination-plan")
	carol, found := env.registry.ResolveUser("carol")
	require.True(t, found)
	replyTo(alice, coordination, plan, "Following the other thread by participating")
	replyTo(carol, coordination, plan, "Second private thread in this batch")
	first := reply(bob, "Private reply content")
	second := reply(bob, "Private append content")
	require.True(t, second.Appended)
	require.Equal(t, first.CommentID, second.CommentID)
	require.NoError(t, producer.NotifyForumReply(ctx, second.Intent))
	messages, err := env.cd.Queries().SystemListPendingEmails(ctx, db.SystemListPendingEmailsParams{Limit: 10})
	require.NoError(t, err)
	require.Empty(t, messages, "optional replies must wait for the batch, including replayed activity")

	worker := newWorker(queries, provider)
	now = time.Date(2026, 11, 21, 10, 14, 59, 0, time.UTC)
	require.NoError(t, worker.ProcessForumReplyEmailWindow(ctx))
	require.Empty(t, provider.Messages)
	now = now.Add(time.Second)
	require.NoError(t, worker.ProcessForumReplyEmailWindow(ctx))
	require.Len(t, provider.Messages, 1)
	message := provider.Messages[0]
	require.Equal(t, "alice-batch@example.test", message.To.Address)
	require.Contains(t, message.Subject, "Forum reply summary")
	require.Contains(t, message.Text, "2 new replies")
	require.Contains(t, message.HTML, "2 new replies")
	require.Contains(t, message.Text, "1 new reply")
	require.NotContains(t, message.Text, "Second private thread in this batch")
	require.NotContains(t, string(message.Raw), "Private reply content")
	require.NotContains(t, string(message.Raw), "Private append content")
	require.NotContains(t, string(message.Raw), "Staff Room")

	require.NoError(t, producer.NotifyForumReply(ctx, second.Intent))
	now = now.Add(15 * time.Minute)
	require.NoError(t, newWorker(queries, provider).ProcessForumReplyEmailWindow(ctx))
	require.Len(t, provider.Messages, 1, "a replay in a later window must not send again")

	reply(bob, "Old activity across an opt-out")
	require.NoError(t, aliceCD.SaveEmail(alice, false, true))
	require.NoError(t, aliceCD.SaveEmail(alice, true, true))
	now = now.Add(15 * time.Minute)
	require.NoError(t, worker.ProcessForumReplyEmailWindow(ctx))
	require.Len(t, provider.Messages, 1, "re-enabling mail does not release the old generation")

	cfg.NotificationsEnabled = false
	reply(bob, "New activity while eligible for email only")
	cfg.NotificationsEnabled = true
	now = now.Add(15 * time.Minute)
	require.NoError(t, worker.ProcessForumReplyEmailWindow(ctx))
	require.Len(t, provider.Messages, 2)
	require.Contains(t, provider.Messages[1].Text, "1 new reply")

	reply(bob, "Reply before Alice unsubscribes from email")
	subs, err := queries.ListSubscriptionsByUser(ctx, alice)
	require.NoError(t, err)
	for _, sub := range subs {
		if sub.Method == "email" {
			require.NoError(t, aliceCD.DeleteSubscription(alice, sub.ID))
		}
	}
	now = now.Add(15 * time.Minute)
	require.NoError(t, worker.ProcessForumReplyEmailWindow(ctx))
	require.Len(t, provider.Messages, 2, "unsubscribe removes delivery eligibility while the window is pending")
	require.NoError(t, common.EnsureSubscriptionIdempotent(ctx, queries, alice, common.ThreadSubscriptionPattern(topic, thread, true), "email"))

	reply(bob, "Reply before a transient preparation failure")
	now = now.Add(15 * time.Minute)
	lookupFailure := errors.New("simulated lookup failure")
	unavailable := newWorker(failedReplyBatchRecipientLookup{Querier: queries, err: lookupFailure}, provider)
	require.ErrorIs(t, unavailable.ProcessForumReplyEmailWindow(ctx), lookupFailure)
	require.NoError(t, newWorker(queries, provider).ProcessForumReplyEmailWindow(ctx))
	require.Len(t, provider.Messages, 2, "a new worker must respect the active lease")
	now = now.Add(2 * time.Minute)
	require.NoError(t, newWorker(queries, provider).ProcessForumReplyEmailWindow(ctx))
	require.Len(t, provider.Messages, 3, "preparation recovers after lease expiry")

	reply(bob, "Reply before an uncertain provider result")
	now = now.Add(15 * time.Minute)
	uncertain := &uncertainReplyBatchProvider{Provider: provider}
	require.ErrorContains(t, newWorker(queries, uncertain).ProcessForumReplyEmailWindow(ctx), "uncertain outcome")
	require.Len(t, provider.Messages, 4)
	now = now.Add(15 * time.Minute)
	require.NoError(t, newWorker(queries, provider).ProcessForumReplyEmailWindow(ctx))
	require.Len(t, provider.Messages, 4, "an ambiguous provider result must not send the message twice")

	reply(bob, "Reply before thread grant is revoked")
	var grantID int32
	require.NoError(t, env.db.QueryRowContext(ctx, "SELECT id FROM grants WHERE section = 'privateforum_thread' AND action = 'view' AND item_id = ? AND user_id = ?", thread, alice).Scan(&grantID))
	require.NoError(t, aliceCD.RevokeForumThread(grantID))
	now = now.Add(15 * time.Minute)
	require.NoError(t, worker.ProcessForumReplyEmailWindow(ctx))
	require.Len(t, provider.Messages, 4, "a revoked private thread must not reach the provider")
}

type failedReplyBatchRecipientLookup struct {
	db.Querier
	err error
}

func (q failedReplyBatchRecipientLookup) SystemGetUserByID(context.Context, int32) (*db.SystemGetUserByIDRow, error) {
	return nil, q.err
}

type uncertainReplyBatchProvider struct{ *mock.Provider }

func (p *uncertainReplyBatchProvider) Send(ctx context.Context, to mail.Address, raw []byte) error {
	if err := p.Provider.Send(ctx, to, raw); err != nil {
		return err
	}
	return errors.New("simulated connection failure after acceptance")
}
