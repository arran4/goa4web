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

	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/opengraph"
	"github.com/arran4/goa4web/workers"
	"github.com/arran4/goa4web/workers/externallinkworker"
	"github.com/stretchr/testify/require"
)

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
