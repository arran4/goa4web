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
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers/images"
	"github.com/arran4/goa4web/internal/db"
	intimages "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/upload"
	"github.com/arran4/goa4web/workers"
	"github.com/gorilla/mux"
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

func TestScenarioReplacedImageClaimCannotReplaceServedBytes(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	q := env.cd.Queries()
	queue := jobs.NewService(q)
	var stale, current bytes.Buffer
	require.NoError(t, png.Encode(&stale, image.NewGray(image.Rect(0, 0, 1, 1))))
	require.NoError(t, png.Encode(&current, image.NewGray(image.Rect(0, 0, 2, 1))))

	cache := &publicationTestCache{
		files: make(map[string][]byte), delayed: stale.Bytes(),
		started: make(chan struct{}), release: make(chan struct{}),
	}
	// providerName identifies the cache controlled by this scenario.
	const providerName = "scenario-image-lease-publication"
	upload.RegisterProvider(providerName, func(*config.RuntimeConfig) upload.Provider { return cache })
	env.server.Config.ImageCacheProvider = providerName
	env.server.Config.ImageThumbnailSizes = "128x128,64x64"
	router := mux.NewRouter()
	images.RegisterRoutes(router, env.server.Config)
	app := common.NewCoreData(ctx, q, env.server.Config, common.WithImageSignKey("scenario-image-signing-key"))
	serve := func(t *testing.T, id string) []byte {
		t.Helper()
		signedURL, err := url.Parse(app.SignCacheURL(id, time.Hour))
		require.NoError(t, err)
		request := httptest.NewRequest(http.MethodGet, signedURL.RequestURI(), nil)
		request = request.WithContext(context.WithValue(request.Context(), consts.KeyCoreData, app))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code)
		return response.Body.Bytes()
	}
	clientFor := func(content []byte) *http.Client {
		return &http.Client{Transport: pendingImageTestTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(bytes.NewReader(content)),
				Header: http.Header{"Content-Type": {"image/png"}},
			}, nil
		})}
	}
	t.Run("Legacy reference still resolves", func(t *testing.T) {
		legacyApp := common.NewCoreData(ctx, q, env.server.Config, common.WithHTTPClient(clientFor(current.Bytes())))
		ref, err := legacyApp.DownloadAndCacheImage("https://example.test/legacy-cache.png")
		require.NoError(t, err)
		require.Equal(t, current.Bytes(), serve(t, strings.TrimPrefix(ref, "cache:")))
	})

	ref, err := env.cd.QueueRemoteImageCache("https://example.test/replaced-claim.png")
	require.NoError(t, err)
	id := strings.TrimPrefix(ref, "cache:")
	require.NoError(t, queue.QueuePendingImageCacheJobs(ctx, id, sql.NullInt32{}))
	first, err := queue.ClaimNext(ctx, "first-image-worker", jobs.JobTypeImageCache)
	require.NoError(t, err)
	require.NotNil(t, first)

	process := func(claim *jobs.ClaimedJob, content []byte) error {
		app := common.NewCoreData(ctx, q, env.server.Config, common.WithHTTPClient(clientFor(content)))
		return app.ProcessClaimedRemoteImageCacheJob(ctx, queue, claim.Job)
	}
	finished := make(chan error, 1)
	go func() { finished <- process(first, stale.Bytes()) }()
	select {
	case <-cache.started:
	case <-ctx.Done():
		t.Fatal("first worker never reached cache storage")
	}

	cancelled, err := queue.AdminCancel(ctx, first.Job.ID)
	require.NoError(t, err)
	require.True(t, cancelled)
	retried, err := queue.AdminRetry(ctx, first.Job.ID)
	require.NoError(t, err)
	require.True(t, retried)
	second, err := queue.ClaimNext(ctx, "second-image-worker", jobs.JobTypeImageCache)
	require.NoError(t, err)
	require.NotNil(t, second)
	require.NoError(t, process(second, current.Bytes()))
	activeThumbnail := strings.TrimSuffix(id, filepath.Ext(id)) + "_thumb_128x128.png"
	publishedThumbnail := serve(t, activeThumbnail)
	close(cache.release)
	select {
	case err := <-finished:
		require.ErrorIs(t, err, jobs.ErrLeaseLost)
	case <-ctx.Done():
		t.Fatal("replaced worker did not finish")
	}

	t.Run("Published objects survive the stale write", func(t *testing.T) {
		require.Equal(t, current.Bytes(), serve(t, id))
		require.Equal(t, publishedThumbnail, serve(t, activeThumbnail))
	})
	t.Run("Evicted thumbnail is restored from the published original", func(t *testing.T) {
		entry, err := q.GetImageCacheEntry(ctx, activeThumbnail)
		require.NoError(t, err)
		key, err := common.ImageCacheObjectKey(activeThumbnail, entry)
		require.NoError(t, err)
		cache.mu.Lock()
		delete(cache.files, key)
		cache.mu.Unlock()
		require.Equal(t, publishedThumbnail, serve(t, activeThumbnail))
	})
	t.Run("On-demand thumbnail uses the published original", func(t *testing.T) {
		source, err := png.Decode(bytes.NewReader(current.Bytes()))
		require.NoError(t, err)
		expected, err := intimages.GenerateThumbnailWithinBounds(source, ".png", env.server.Config.ImageThumbnailGenerator, 64, 64)
		require.NoError(t, err)
		alt := strings.TrimSuffix(id, filepath.Ext(id)) + "_thumb_64x64.png"
		require.Equal(t, expected, serve(t, alt))
	})
}

func TestScenarioImageFailureDiagnosticsOmitSourceURL(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := t.Context()
	q := env.cd.Queries()
	queue := jobs.NewService(q)
	var adminMessage string
	queue.NotifyAdmins = func(_ context.Context, message, _ string) error {
		adminMessage = message
		return nil
	}
	// privateValue represents an embedded capability in a private-forum image URL.
	const privateValue = "example-private-image-capability"
	ref, err := env.cd.QueueRemoteImageCache("https://example.test/private.png?auth=" + privateValue)
	require.NoError(t, err)
	id := strings.TrimPrefix(ref, "cache:")
	_, err = queue.Enqueue(ctx, jobs.EnqueueRequest{
		JobType: jobs.JobTypeImageCache, TargetType: jobs.TargetTypeImageCacheEntry, TargetKey: id, MaxAttempts: 1,
	})
	require.NoError(t, err)
	claim, err := queue.ClaimNext(ctx, "worker", jobs.JobTypeImageCache)
	require.NoError(t, err)
	require.NotNil(t, claim)

	transportFailure := errors.New("upstream connection timed out")
	client := &http.Client{Transport: pendingImageTestTransport(func(*http.Request) (*http.Response, error) {
		return nil, transportFailure
	})}
	app := common.NewCoreData(ctx, q, env.server.Config, common.WithHTTPClient(client))
	fetchErr := app.ProcessClaimedRemoteImageCacheJob(ctx, queue, claim.Job)
	require.ErrorIs(t, fetchErr, transportFailure)
	applied, err := queue.FailOrRetryFenced(ctx, claim, fetchErr)
	require.NoError(t, err)
	require.True(t, applied)

	entry, err := app.ImageCacheEntry(ctx, id)
	require.NoError(t, err)
	job, err := q.SystemGetDurableJobByID(ctx, claim.Job.ID)
	require.NoError(t, err)
	for label, value := range map[string]string{
		"returned error": fetchErr.Error(), "image record": entry.ErrorMessage.String,
		"job record": job.LastError.String, "admin notification": adminMessage,
	} {
		t.Run(label, func(t *testing.T) {
			require.NotEmpty(t, value)
			require.NotContains(t, value, privateValue)
		})
	}
}

type publicationTestCache struct {
	mu      sync.Mutex
	once    sync.Once
	files   map[string][]byte
	delayed []byte
	started chan struct{}
	release chan struct{}
}

func (*publicationTestCache) Check(context.Context) error { return nil }

func (c *publicationTestCache) Write(ctx context.Context, key string, data []byte) error {
	if bytes.Equal(data, c.delayed) {
		c.once.Do(func() { close(c.started) })
		select {
		case <-c.release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.files[key] = bytes.Clone(data)
	return nil
}

func (c *publicationTestCache) Read(_ context.Context, key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if data, ok := c.files[key]; ok {
		return bytes.Clone(data), nil
	}
	return nil, errors.New("missing cache object")
}
