//go:build sqlite || sqlite3

package main

import (
	"context"
	"errors"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/scenario"
	"github.com/stretchr/testify/require"
)

func TestScenarioForumWritesRetainCommittedResultsAfterProjectionFailure(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()
	bob, ok := env.registry.ResolveUser("bob")
	require.True(t, ok)
	alice, ok := env.registry.ResolveUser("alice")
	require.True(t, ok)
	topic := requireScenarioID(t, env.registry, scenario.RefTypeForum, "staff-room")
	workerErr := errors.New("topic projection unavailable")
	notifier := &failingReplyProjection{err: errors.New("notification projection unavailable")}
	cfg := *env.server.Config
	cfg.PrivateForumPostAppendWindow = 0
	app := common.NewCoreData(ctx, failingForumProjectionQueries{Querier: env.cd.Queries(), err: workerErr}, &cfg,
		common.WithForumReplyNotifier(notifier))

	thread, err := app.ForUser(bob).CreateForumThread(ctx, common.CreateForumThreadParams{
		ActorID: bob, TopicID: topic, Text: "Committed opener", LanguageID: 1,
		Private: true, EnforceHandler: true, SynchronousSideEffects: true,
	})
	require.NoError(t, err)
	require.ErrorIs(t, thread.SideEffectError, workerErr)
	require.Positive(t, thread.ThreadID)

	reply, err := app.ForUser(alice).ReplyForumThread(ctx, common.ReplyForumThreadParams{
		ActorID: alice, TopicID: topic, ThreadID: thread.ThreadID, Text: "Committed reply", LanguageID: 1,
		Private: true, EnforceHandler: true, SynchronousSideEffects: true,
	})
	require.NoError(t, err)
	require.ErrorIs(t, reply.SideEffectError, workerErr)
	require.ErrorIs(t, reply.SideEffectError, notifier.err)
	require.Equal(t, []string{reply.ActivityID}, notifier.attempted)

	for id, expected := range map[int32]string{thread.CommentID: "Committed opener", reply.CommentID: "Committed reply"} {
		var body string
		require.NoError(t, env.db.QueryRowContext(ctx, "SELECT text FROM comments WHERE idcomments = ? AND forumthread_id = ?", id, thread.ThreadID).Scan(&body))
		require.Equal(t, expected, body)
	}
}

type failingForumProjectionQueries struct {
	db.Querier
	err error
}

func (q failingForumProjectionQueries) SystemRebuildForumTopicMetaByID(context.Context, int32) error {
	return q.err
}

type failingReplyProjection struct {
	err       error
	attempted []string
}

func (p *failingReplyProjection) NotifyForumReply(_ context.Context, intent common.ForumReplyNotificationIntent) error {
	p.attempted = append(p.attempted, intent.ActivityID)
	return p.err
}
