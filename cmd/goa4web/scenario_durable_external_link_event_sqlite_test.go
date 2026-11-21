//go:build sqlite || sqlite3

package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/workers"
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
