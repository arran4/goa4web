//go:build sqlite || sqlite3

package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"image"
	"image/png"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
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

func TestScenarioImageRecoveryHonorsCancellation(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()
	queue := jobs.NewService(env.cd.Queries())
	ref, err := env.cd.QueueRemoteImageCache("https://example.test/cancelled.png")
	require.NoError(t, err)
	id := strings.TrimPrefix(ref, "cache:")

	require.NoError(t, queue.QueuePendingImageCacheJobs(ctx, id, sql.NullInt32{}))
	rows, err := queue.AdminList(ctx, jobs.StatusPending, jobs.JobTypeImageCache, 20, 0)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	jobID := rows[0].ID

	_, err = env.db.ExecContext(ctx, `CREATE TRIGGER reject_image_cancellation BEFORE UPDATE OF status ON image_cache_entries
		WHEN NEW.status = 'cancelled' BEGIN SELECT RAISE(ABORT, 'injected cache failure'); END`)
	require.NoError(t, err)
	cancelled, err := queue.AdminCancel(ctx, jobID)
	require.ErrorContains(t, err, "injected cache failure")
	require.False(t, cancelled)
	job, err := env.cd.Queries().SystemGetDurableJobByID(ctx, jobID)
	require.NoError(t, err)
	require.Equal(t, jobs.StatusPending, job.Status, "a failed cache update must roll back queue cancellation")
	_, err = env.db.ExecContext(ctx, "DROP TRIGGER reject_image_cancellation")
	require.NoError(t, err)

	cancelled, err = queue.AdminCancel(ctx, jobID)
	require.NoError(t, err)
	require.True(t, cancelled)
	require.NoError(t, queue.QueuePendingImageCacheJobs(ctx, "", sql.NullInt32{}))
	rows, err = queue.AdminList(ctx, jobs.StatusPending, jobs.JobTypeImageCache, 20, 0)
	require.NoError(t, err)
	require.Empty(t, rows, "recovery must not queue a cancelled image")

	// A worker that was interrupted during its fetch may still record a failure.
	require.NoError(t, env.cd.Queries().RecordImageCacheFetchFailure(ctx, db.RecordImageCacheFetchFailureParams{
		ID: id, RetryCount: 3, RetryCount_2: 3,
	}))
	entry, err := env.cd.ImageCacheEntry(ctx, id)
	require.NoError(t, err)
	require.Equal(t, jobs.StatusCancelled, entry.Status)

	retried, err := queue.AdminRetry(ctx, jobID)
	require.NoError(t, err)
	require.True(t, retried)
	entry, err = env.cd.ImageCacheEntry(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "pending", entry.Status)

	cancelled, err = queue.AdminCancel(ctx, jobID)
	require.NoError(t, err)
	require.True(t, cancelled)
	purged, err := queue.AdminPurge(ctx, jobs.StatusCancelled)
	require.NoError(t, err)
	require.Equal(t, int64(1), purged)
	require.NoError(t, queue.QueuePendingImageCacheJobs(ctx, "", sql.NullInt32{}))
	rows, err = queue.AdminList(ctx, jobs.StatusPending, jobs.JobTypeImageCache, 20, 0)
	require.NoError(t, err)
	require.Empty(t, rows, "pruning the job history must not undo cancellation")
}

func TestScenarioCancelledImageDoesNotPublishInFlightSuccess(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()
	q := env.cd.Queries()
	queue := jobs.NewService(q)

	ref, err := env.cd.QueueRemoteImageCache("https://example.test/cancel-during-fetch.png")
	require.NoError(t, err)
	id := strings.TrimPrefix(ref, "cache:")
	require.NoError(t, queue.QueuePendingImageCacheJobs(ctx, id, sql.NullInt32{}))
	claim, err := queue.ClaimNext(ctx, "image-worker", jobs.JobTypeImageCache)
	require.NoError(t, err)
	require.NotNil(t, claim)

	var pngBytes bytes.Buffer
	require.NoError(t, png.Encode(&pngBytes, image.NewGray(image.Rect(0, 0, 1, 1))))
	started := make(chan struct{})
	complete := make(chan struct{})
	client := &http.Client{Transport: pendingImageTestTransport(func(request *http.Request) (*http.Response, error) {
		close(started)
		select {
		case <-complete:
			return &http.Response{
				StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(bytes.NewReader(pngBytes.Bytes())),
				Header: http.Header{"Content-Type": {"image/png"}},
			}, nil
		case <-request.Context().Done():
			return nil, request.Context().Err()
		}
	})}

	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	app := common.NewCoreData(workerCtx, q, env.server.Config, common.WithHTTPClient(client))
	result := make(chan error, 1)
	go func() { result <- app.ProcessClaimedRemoteImageCacheJob(workerCtx, queue, claim.Job) }()

	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("claimed image fetch did not start")
	}
	cancelled, err := queue.AdminCancel(ctx, claim.Job.ID)
	require.NoError(t, err)
	require.True(t, cancelled)
	close(complete)
	select {
	case err := <-result:
		require.ErrorIs(t, err, jobs.ErrLeaseLost)
	case <-time.After(3 * time.Second):
		t.Fatal("claimed image fetch did not finish")
	}

	entry, err := app.ImageCacheEntry(ctx, id)
	require.NoError(t, err)
	require.Equal(t, jobs.StatusCancelled, entry.Status, "a stale fetch must not publish after cancellation")
}
