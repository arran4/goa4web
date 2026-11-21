//go:build sqlite || sqlite3

package main

import (
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/scenario"
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
