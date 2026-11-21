//go:build sqlite || sqlite3

package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/workers"
	"github.com/stretchr/testify/require"
)

func TestScenarioStandaloneWorkerRecoversPendingImage(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()
	q := env.cd.Queries()

	// imageURL identifies the pending image whose enqueue was interrupted.
	const imageURL = "https://example.test/recover-pending.png"
	ref, err := env.cd.QueueRemoteImageCache(imageURL)
	require.NoError(t, err)

	var requests atomic.Int32
	client := &http.Client{Transport: pendingImageTestTransport(func(request *http.Request) (*http.Response, error) {
		if request.URL.String() == imageURL {
			requests.Add(1)
		}
		return nil, errors.New("fixture fetch failed")
	})}
	queue := jobs.NewService(q)
	group := workers.StartDurableWorkers(ctx, q, env.server.Config, queue, "image-recovery",
		workers.WithCoreOptions(common.WithHTTPClient(client)))
	defer func() {
		group.Cancel()
		waitCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		require.NoError(t, group.Wait(waitCtx))
	}()

	require.Eventually(t, func() bool {
		rows, err := queue.AdminList(ctx, jobs.StatusPending, jobs.JobTypeImageCache, 20, 0)
		if err != nil {
			return false
		}
		for _, row := range rows {
			if row.TargetKey == strings.TrimPrefix(ref, "cache:") && row.Attempts == 1 {
				return true
			}
		}
		return false
	}, 4*time.Second, 50*time.Millisecond, "the recovered fetch must schedule its next retry through the durable queue")
	before, err := queue.AdminList(ctx, jobs.StatusPending, jobs.JobTypeImageCache, 20, 0)
	require.NoError(t, err)
	require.Len(t, before, 1)
	require.NoError(t, queue.QueuePendingImageCacheJobs(ctx, strings.TrimPrefix(ref, "cache:"), sql.NullInt32{}))
	require.NoError(t, queue.QueuePendingImageCacheJobs(ctx, "", sql.NullInt32{}))
	after, err := queue.AdminList(ctx, jobs.StatusPending, jobs.JobTypeImageCache, 20, 0)
	require.NoError(t, err)
	require.Len(t, after, 1)
	require.Equal(t, before[0].ID, after[0].ID)
	require.Equal(t, before[0].RunAfter, after[0].RunAfter)
	require.Equal(t, int32(1), requests.Load())
}

type pendingImageTestTransport func(*http.Request) (*http.Response, error)

func (f pendingImageTestTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
