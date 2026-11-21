//go:build sqlite || sqlite3

package main

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/opengraph"
	"github.com/arran4/goa4web/internal/scenario"
	"github.com/arran4/goa4web/workers"
	"github.com/arran4/goa4web/workers/externallinkworker"
	"github.com/stretchr/testify/require"
)

func TestScenarioForumOpeningCommentCommitsWithObservedLink(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()
	q := env.cd.Queries()
	queue := jobs.NewService(q)
	aliceID, ok := env.registry.ResolveUser("alice")
	require.True(t, ok)
	topicID := requireScenarioID(t, env.registry, scenario.RefTypeForum, "staff-room")
	app := common.NewCoreData(ctx, q, env.server.Config, common.WithJobService(queue)).ForUser(aliceID)
	targetURL := "https://example.test/from-committed-opening-comment"
	params := common.CreateForumThreadParams{
		ActorID: aliceID, TopicID: topicID, LanguageID: 1,
		Text: "[a=" + targetURL + "]Opening post link", Private: true, EnforceHandler: true,
	}

	threads, err := q.AdminListPrivateForumThreads(ctx, db.AdminListPrivateForumThreadsParams{Limit: 200})
	require.NoError(t, err)
	_, err = env.db.ExecContext(ctx, `CREATE TRIGGER reject_observed_link BEFORE INSERT ON durable_jobs
		WHEN NEW.job_type = 'external_link' BEGIN SELECT RAISE(ABORT, 'injected observed link failure'); END`)
	require.NoError(t, err)
	result, err := app.CreateForumThread(ctx, params)
	require.ErrorContains(t, err, "injected observed link failure")
	require.Nil(t, result)
	after, err := q.AdminListPrivateForumThreads(ctx, db.AdminListPrivateForumThreadsParams{Limit: 200})
	require.NoError(t, err)
	require.Len(t, after, len(threads))

	_, err = env.db.ExecContext(ctx, "DROP TRIGGER reject_observed_link")
	require.NoError(t, err)
	result, err = app.CreateForumThread(ctx, params)
	require.NoError(t, err)
	require.NotZero(t, result.CommentID)
	claim, err := queue.ClaimNext(ctx, "standalone-metadata", jobs.JobTypeExternalLink)
	require.NoError(t, err)
	require.NotNil(t, claim)
	require.Equal(t, targetURL, claim.Job.TargetKey)
}

func TestScenarioForumReplyAppendAndEditCommitWithObservedLinks(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()
	q := env.cd.Queries()
	queue := jobs.NewService(q)
	alice, ok := env.registry.ResolveUser("alice")
	require.True(t, ok)
	topic := requireScenarioID(t, env.registry, scenario.RefTypeForum, "staff-room")
	thread := requireScenarioID(t, env.registry, scenario.RefTypeThread, "staff-welcome")
	cfg := *env.server.Config
	cfg.PrivateForumPostAppendWindow = 0
	app := func() *common.CoreData {
		return common.NewCoreData(ctx, q, &cfg, common.WithJobService(queue)).ForUser(alice)
	}
	now := time.Now().UTC().Truncate(time.Second)
	read := func(t *testing.T) []*db.GetCommentsByThreadIdForUserRow {
		t.Helper()
		comments, err := app().ThreadComments(thread)
		require.NoError(t, err)
		return comments
	}
	tryMutation := func(t *testing.T, url string, mutate func(string) error) {
		t.Helper()
		before := read(t)
		_, err := env.db.ExecContext(ctx, `CREATE TRIGGER reject_link_write BEFORE INSERT ON durable_jobs
			WHEN NEW.job_type = 'external_link' BEGIN SELECT RAISE(ABORT, 'injected observed link failure'); END`)
		require.NoError(t, err)
		body := "[a=" + url + "]new link"
		require.ErrorContains(t, mutate(body), "injected observed link failure")
		require.Equal(t, before, read(t), "the source cannot change without its link observation")
		_, err = env.db.ExecContext(ctx, "DROP TRIGGER reject_link_write")
		require.NoError(t, err)
		require.NoError(t, mutate(body))
		claim, err := queue.ClaimNext(ctx, "standalone-metadata", jobs.JobTypeExternalLink)
		require.NoError(t, err)
		require.NotNil(t, claim)
		require.Equal(t, url, claim.Job.TargetKey)
		require.Equal(t, sql.NullInt32{Int32: alice, Valid: true}, claim.Job.CreatedByUserID)
		completed, err := queue.CompleteFenced(ctx, claim)
		require.NoError(t, err)
		require.True(t, completed)
	}

	var replyID int32
	t.Run("Reply", func(t *testing.T) {
		tryMutation(t, "https://example.test/from-forum-reply", func(body string) error {
			// This link exceeds the durable key width and must not abort the post.
			body += "\n[a=https://example.test/" + strings.Repeat("x", 512) + "]Unqueueable link"
			reply, err := app().ReplyForumThread(ctx, common.ReplyForumThreadParams{
				ActorID: alice, TopicID: topic, ThreadID: thread, LanguageID: 1,
				Text: body, Private: true, EnforceHandler: true, At: now.Add(time.Second),
			})
			if err == nil {
				require.False(t, reply.Appended)
				replyID = reply.CommentID
			}
			return err
		})
	})
	require.NotZero(t, replyID)
	cfg.PrivateForumPostAppendWindow = 60
	t.Run("Append", func(t *testing.T) {
		tryMutation(t, "https://example.test/from-forum-append", func(body string) error {
			reply, err := app().ReplyForumThread(ctx, common.ReplyForumThreadParams{
				ActorID: alice, TopicID: topic, ThreadID: thread, LanguageID: 1,
				Text: body, Private: true, EnforceHandler: true, At: now.Add(2 * time.Second),
			})
			if err == nil {
				require.True(t, reply.Appended)
				require.Equal(t, replyID, reply.CommentID)
				require.NotZero(t, reply.AppendID)
			}
			return err
		})
	})
	t.Run("Edit", func(t *testing.T) {
		tryMutation(t, "https://example.test/from-forum-edit", func(body string) error {
			return app().EditForumCommentAction(ctx, common.EditForumCommentParams{
				ActorID: alice, CommentID: replyID, LanguageID: 1, Text: body,
			})
		})
	})
}

func TestScenarioLegacyExternalLinkEventsUseDurableQueue(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()
	queue := jobs.NewService(env.cd.Queries())
	now := time.Now().UTC()
	queue.Now = func() time.Time { return now }

	// retryURL identifies a link already waiting for its next durable attempt.
	const retryURL = "https://example.test/already-retrying"
	created, err := queue.Enqueue(ctx, jobs.EnqueueRequest{
		JobType: jobs.JobTypeExternalLink, TargetType: jobs.JobTypeExternalLink, TargetKey: retryURL,
	})
	require.NoError(t, err)
	require.True(t, created)
	claim, err := queue.ClaimNext(ctx, "test", jobs.JobTypeExternalLink)
	require.NoError(t, err)
	require.NotNil(t, claim)
	_, err = queue.FailOrRetryFenced(ctx, claim, errors.New("temporary failure"))
	require.NoError(t, err)
	before, err := env.cd.Queries().SystemGetDurableJobByID(ctx, claim.Job.ID)
	require.NoError(t, err)

	bus := eventbus.NewBus()
	group := workers.Start(ctx, env.cd.Queries(), nil, nil, env.server.Config, bus, workers.WithJobService(queue))
	defer func() {
		group.Cancel()
		waitCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		require.NoError(t, group.Wait(waitCtx))
	}()

	// newURL identifies an older task producer's link with no existing job.
	const newURL = "https://example.test/from-legacy-task"
	require.Eventually(t, func() bool {
		_ = bus.Publish(eventbus.TaskEvent{
			Outcome: eventbus.TaskOutcomeSuccess,
			Data:    map[string]any{"Body": "[a=" + retryURL + "]existing\n[a=" + newURL + "]new"},
		})
		list, err := queue.AdminList(ctx, "", jobs.JobTypeExternalLink, 20, 0)
		if err != nil {
			return false
		}
		for _, row := range list {
			if row.TargetKey == newURL {
				return true
			}
		}
		return false
	}, 6*time.Second, 100*time.Millisecond, "the legacy producer must feed the same durable worker")
	retained, err := env.cd.Queries().SystemGetDurableJobByID(ctx, claim.Job.ID)
	require.NoError(t, err)
	require.Equal(t, int32(1), retained.Attempts)
	require.Equal(t, before.RunAfter, retained.RunAfter, "observing a link must preserve its scheduled retry")

	// recordedURL represents an already projected observation with no active job left.
	const recordedURL = "https://example.test/already-recorded"
	require.NoError(t, bus.Publish(eventbus.TaskEvent{
		Outcome: eventbus.TaskOutcomeSuccess,
		Data: map[string]any{
			"Body":                            "[a=" + recordedURL + "]recorded",
			eventbus.ExternalLinksRecordedKey: true,
		},
	}))
	drainCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	require.NoError(t, bus.Shutdown(drainCtx))
	rows, err := queue.AdminList(ctx, "", jobs.JobTypeExternalLink, 20, 0)
	require.NoError(t, err)
	for _, row := range rows {
		require.NotEqual(t, recordedURL, row.TargetKey, "dispatch must not recreate an observation recorded with the source")
	}
}

func TestScenarioDurableExternalLinkHonorsHTTPFailurePolicy(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()
	q := env.cd.Queries()
	queue := jobs.NewService(q)
	now := time.Date(2026, time.November, 21, 12, 0, 0, 0, time.UTC)
	queue.Now = func() time.Time { return now }

	for _, tc := range []struct {
		name   string
		status int
		header string
		delay  time.Duration
	}{
		{"rate-limit", http.StatusTooManyRequests, "90", 90 * time.Second},
		{"service-unavailable", http.StatusServiceUnavailable, now.Add(time.Hour).Format(http.TimeFormat), 5 * time.Minute},
		{"overflowing-retry", http.StatusTooManyRequests, "999999999999999999999999", 5 * time.Minute},
		{"server-error", http.StatusBadGateway, "", 10 * time.Second},
		{"invalid-retry", http.StatusTooManyRequests, "-10", 10 * time.Second},
		{"gone", http.StatusGone, "", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := "https://example.test/" + tc.name
			created, err := queue.Enqueue(ctx, jobs.EnqueueRequest{JobType: jobs.JobTypeExternalLink, TargetType: jobs.JobTypeExternalLink, TargetKey: target})
			require.NoError(t, err)
			require.True(t, created)
			claim, err := queue.ClaimNext(ctx, "external-link-http", jobs.JobTypeExternalLink)
			require.NoError(t, err)
			require.NotNil(t, claim)
			require.Equal(t, target, claim.Job.TargetKey)
			client := &http.Client{Transport: pendingImageTestTransport(func(_ *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: tc.status,
					Header:     http.Header{"Retry-After": {tc.header}},
					Body:       io.NopCloser(strings.NewReader("<html><head><title>Upstream error page</title></head></html>")),
				}, nil
			})}
			result := externallinkworker.ProcessClaimedExternalLink(ctx, queue, q, env.server.Config, claim.Job, client)
			var status *opengraph.HTTPStatusError
			require.ErrorAs(t, result, &status)
			require.Equal(t, tc.status, status.StatusCode)
			applied, err := queue.FailOrRetryFenced(ctx, claim, result)
			require.NoError(t, err)
			require.True(t, applied)
			record, err := q.GetExternalLink(ctx, target)
			require.NoError(t, err)
			require.Empty(t, record.CardTitle.String)
			job, err := q.SystemGetDurableJobByID(ctx, claim.Job.ID)
			require.NoError(t, err)
			if tc.delay != 0 {
				require.Equal(t, jobs.StatusPending, job.Status)
				require.True(t, now.Add(tc.delay).Equal(job.RunAfter), "run after %s", job.RunAfter)
				_, err = queue.AdminCancel(ctx, job.ID)
				require.NoError(t, err)
				return
			}
			require.Equal(t, jobs.StatusFailed, job.Status)
			require.Equal(t, int32(1), job.Attempts)
			now = now.Add(48 * time.Hour)
			require.NoError(t, queue.QueueObservedExternalLink(ctx, target, sql.NullInt32{}))
			next, err := queue.ClaimNext(ctx, "no-automatic-retry", jobs.JobTypeExternalLink)
			require.NoError(t, err)
			require.Nil(t, next)
			retried, err := queue.AdminRetry(ctx, job.ID)
			require.NoError(t, err)
			require.True(t, retried)
		})
	}
}

func TestScenarioCancelledExternalLinkClaimCannotPublishMetadata(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()
	q := env.cd.Queries()
	queue := jobs.NewService(q)

	// target identifies the queued link cancelled during its fetch.
	const target = "https://example.test/cancel-during-link-fetch"
	created, err := queue.Enqueue(ctx, jobs.EnqueueRequest{JobType: jobs.JobTypeExternalLink, TargetType: jobs.JobTypeExternalLink, TargetKey: target})
	require.NoError(t, err)
	require.True(t, created)
	claim, err := queue.ClaimNext(ctx, "external-link-cancellation", jobs.JobTypeExternalLink)
	require.NoError(t, err)
	require.NotNil(t, claim)
	client := &http.Client{Transport: pendingImageTestTransport(func(req *http.Request) (*http.Response, error) {
		cancelled, err := queue.AdminCancel(req.Context(), claim.Job.ID)
		require.NoError(t, err)
		require.True(t, cancelled)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("<html><head><title>Stale metadata</title></head></html>")),
		}, nil
	})}
	result := externallinkworker.ProcessClaimedExternalLink(ctx, queue, q, env.server.Config, claim.Job, client)
	require.ErrorIs(t, result, jobs.ErrLeaseLost)
	record, err := q.GetExternalLink(ctx, target)
	require.NoError(t, err)
	require.Empty(t, record.CardTitle.String)
}
