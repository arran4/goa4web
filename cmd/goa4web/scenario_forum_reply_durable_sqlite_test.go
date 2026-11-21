//go:build sqlite || sqlite3

package main

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/scenario"
	"github.com/arran4/goa4web/workers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenarioForumReplyIntentSurvivesLostDispatch(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()
	bob, ok := env.registry.ResolveUser("bob")
	require.True(t, ok)
	topic := requireScenarioID(t, env.registry, scenario.RefTypeForum, "staff-room")
	thread := requireScenarioID(t, env.registry, scenario.RefTypeThread, "staff-welcome")
	q := env.cd.Queries()
	producer := common.NewCoreData(ctx, q, env.server.Config, common.WithJobService(jobs.NewService(q)))

	first := common.ReplyForumThreadParams{
		ActorID: bob, TopicID: topic, ThreadID: thread, Text: "Ordinary reply", LanguageID: 1,
		Private: true, EnforceHandler: true, At: time.Now().UTC().Add(-time.Minute),
	}
	replyAsBob := producer.ForUser(bob)
	blockIntent := func() {
		t.Helper()
		_, err := env.db.ExecContext(ctx, `CREATE TRIGGER reject_reply_intent BEFORE INSERT ON durable_jobs
			WHEN NEW.job_type = 'forum_reply_notification'
			BEGIN SELECT RAISE(ABORT, 'queue unavailable'); END`)
		require.NoError(t, err)
	}
	unblockIntent := func() {
		t.Helper()
		_, err := env.db.ExecContext(ctx, "DROP TRIGGER reject_reply_intent")
		require.NoError(t, err)
	}
	before, err := replyAsBob.ThreadComments(thread)
	require.NoError(t, err)
	blockIntent()
	failed, err := replyAsBob.ReplyForumThread(ctx, first)
	require.ErrorContains(t, err, "queue unavailable")
	require.Nil(t, failed)
	after, err := producer.ForUser(bob).ThreadComments(thread)
	require.NoError(t, err)
	require.Len(t, after, len(before))
	unblockIntent()

	ordinary, err := producer.ForUser(bob).ReplyForumThread(ctx, first)
	require.NoError(t, err)
	require.False(t, ordinary.Appended)

	second := first
	second.Text = "Appended reply"
	second.At = second.At.Add(time.Second)
	blockIntent()
	failed, err = producer.ForUser(bob).ReplyForumThread(ctx, second)
	require.ErrorContains(t, err, "queue unavailable")
	require.Nil(t, failed)
	after, err = producer.ForUser(bob).ThreadComments(thread)
	require.NoError(t, err)
	require.Len(t, after, len(before)+1)
	require.Equal(t, first.Text, after[len(after)-1].Text.String)
	failedAppends, err := q.SystemListCommentAppendsByCommentID(ctx, ordinary.CommentID)
	require.NoError(t, err)
	require.Empty(t, failedAppends)
	unblockIntent()

	appended, err := producer.ForUser(bob).ReplyForumThread(ctx, second)
	require.NoError(t, err)
	require.True(t, appended.Appended)
	require.Equal(t, ordinary.CommentID, appended.CommentID)
	require.NotEqual(t, ordinary.ActivityID, appended.ActivityID)

	// Start with a new service and no producer wakeups, bus, or notifier.
	worker := jobs.NewService(q)
	for _, result := range []*common.ReplyForumThreadResult{ordinary, appended} {
		claimed, err := worker.ClaimNext(ctx, "restarted-worker", jobs.JobTypeForumReplyNotification)
		require.NoError(t, err)
		require.NotNil(t, claimed, "committed activity %s must remain claimable", result.ActivityID)
		require.Equal(t, int64(result.CommentID), claimed.Job.TargetID)
		require.Equal(t, result.ActivityID, claimed.Job.TargetKey)
		require.Equal(t, result.Intent.ActivityPosition, claimed.Job.TargetRevision)
		require.Equal(t, bob, claimed.Job.CreatedByUserID.Int32)
	}
}

func TestScenarioForumReplyWorkerDeliversQueuedActivities(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()
	bob, ok := env.registry.ResolveUser("bob")
	require.True(t, ok)
	alice, ok := env.registry.ResolveUser("alice")
	require.True(t, ok)
	topic := requireScenarioID(t, env.registry, scenario.RefTypeForum, "staff-room")
	thread := requireScenarioID(t, env.registry, scenario.RefTypeThread, "staff-welcome")
	require.NoError(t, env.cd.SubscribeForum(ctx, common.SubscribeForumParams{
		ActorID: alice, TopicID: topic, ThreadID: thread,
	}))
	q := env.cd.Queries()
	require.NoError(t, env.cd.ForUser(alice).SaveEmail(alice, true, true))
	require.NoError(t, common.EnsureSubscriptionIdempotent(ctx, q, alice, common.ThreadSubscriptionPattern(topic, thread, true), "email"))
	producer := common.NewCoreData(ctx, q, env.server.Config, common.WithJobService(jobs.NewService(q)))
	reply := common.ReplyForumThreadParams{
		ActorID: bob, TopicID: topic, ThreadID: thread, Text: "Durably delivered reply", LanguageID: 1,
		Private: true, EnforceHandler: true, At: time.Now().UTC().Add(-time.Minute),
	}
	first, err := producer.ForUser(bob).ReplyForumThread(ctx, reply)
	require.NoError(t, err)
	reply.Text = "Durably delivered append"
	reply.At = reply.At.Add(time.Second)
	second, err := producer.ForUser(bob).ReplyForumThread(ctx, reply)
	require.NoError(t, err)
	require.True(t, second.Appended)
	require.True(t, first.Intent.Queued)
	require.True(t, second.Intent.Queued)
	require.NoError(t, env.cd.ForUser(alice).ReadForumThread(ctx, common.ReadForumThreadParams{
		ActorID: alice, ThreadID: thread, LastActivityID: first.ActivityID,
	}))

	legacy := notifications.New(notifications.WithQueries(q), notifications.WithConfig(env.server.Config))
	require.NoError(t, legacy.NotifyForumReply(ctx, first.Intent))
	require.NoError(t, legacy.ProcessEvent(ctx, eventbus.TaskEvent{
		Outcome: eventbus.TaskOutcomeSuccess,
		Data:    map[string]any{common.ForumReplyIntentEventKey: second.Intent},
	}, nil))
	before, err := q.ListNotificationsForLister(ctx, db.ListNotificationsForListerParams{ListerID: alice, Limit: 50})
	require.NoError(t, err)
	for _, notice := range before {
		require.NotEqual(t, first.Intent.ActivityPosition, notice.ActivityPosition.Int64)
		require.NotEqual(t, second.Intent.ActivityPosition, notice.ActivityPosition.Int64)
	}

	queue := jobs.NewService(q)
	group := workers.StartDurableWorkers(ctx, q, env.server.Config, queue, "restarted")
	stop := func(group *workers.Group) {
		t.Helper()
		group.Cancel()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		require.NoError(t, group.Wait(shutdown))
	}
	defer func() { stop(group) }()
	waitForCompleted := func(count int) {
		t.Helper()
		require.EventuallyWithT(t, func(c *assert.CollectT) {
			completed, err := queue.AdminList(ctx, jobs.StatusCompleted, jobs.JobTypeForumReplyNotification, 20, 0)
			assert.NoError(c, err)
			assert.Len(c, completed, count)
		}, 7*time.Second, 50*time.Millisecond)
	}
	waitForCompleted(2)

	// Replay the append after successful delivery, just as a recovered lease may
	// repeat delivery after notification writes but before the job is finalized.
	queued, err := queue.Enqueue(ctx, jobs.EnqueueRequest{
		JobType: jobs.JobTypeForumReplyNotification, TargetType: "comment",
		TargetID: int64(second.CommentID), TargetKey: second.ActivityID, TargetRevision: second.Intent.ActivityPosition,
		CreatedByUserID: sql.NullInt32{Int32: bob, Valid: true},
	})
	require.NoError(t, err)
	require.True(t, queued)
	waitForCompleted(3)

	list, err := q.ListNotificationsForLister(ctx, db.ListNotificationsForListerParams{ListerID: alice, Limit: 50})
	require.NoError(t, err)
	link := fmt.Sprintf("/private/topic/%d/thread/%d#comment-%d", topic, thread, first.CommentID)
	for _, outcome := range []*common.ReplyForumThreadResult{first, second} {
		found := 0
		for _, notice := range list {
			if notice.ActivityPosition.Int64 == outcome.Intent.ActivityPosition {
				found++
				require.Contains(t, notice.Message.String, "bob replied to Staff Room")
				require.Contains(t, notice.Link.String, link)
				require.Equal(t, outcome == first, notice.ReadAt.Valid, "only the acknowledged activity is read")
			}
		}
		require.Equal(t, 1, found, "activity %s", outcome.ActivityID)
	}

	var emailMembers int
	require.NoError(t, env.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM forum_reply_email_activities WHERE recipient_id = ? AND thread_id = ?", alice, thread).Scan(&emailMembers))
	require.Equal(t, 2, emailMembers, "replaying an activity must not repeat its email membership")

	stop(group)
	reply.Text = "Reply before access was revoked"
	reply.At = reply.At.Add(time.Second)
	revoked, err := producer.ForUser(bob).ReplyForumThread(ctx, reply)
	require.NoError(t, err)

	var grantID int32
	require.NoError(t, env.db.QueryRowContext(ctx, "SELECT id FROM grants WHERE section = 'privateforum_thread' AND action = 'view' AND item_id = ? AND user_id = ?", thread, alice).Scan(&grantID))
	require.NoError(t, env.cd.RevokeForumThread(grantID))
	group = workers.StartDurableWorkers(ctx, q, env.server.Config, queue, "after-revocation")
	waitForCompleted(4)

	// Inspect persisted delivery because the inbox itself hides inaccessible entries.
	var alerts, emails int
	require.NoError(t, env.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM notifications WHERE users_idusers = ? AND activity_position = ?", alice, revoked.Intent.ActivityPosition).Scan(&alerts))
	require.NoError(t, env.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM forum_reply_email_activities WHERE recipient_id = ? AND activity_position = ?", alice, revoked.Intent.ActivityPosition).Scan(&emails))
	require.Zero(t, alerts)
	require.Zero(t, emails)
}
