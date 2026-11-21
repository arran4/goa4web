//go:build sqlite

package jobs_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/opengraph"
	"github.com/arran4/goa4web/workers/searchworker"
	"github.com/stretchr/testify/require"
)

func openSharedSQLiteFile(t *testing.T) (*sql.DB, db.Querier) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "jobs.sqlite")
	sdb, err := sql.Open("sqlite", "file:"+dbPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sdb.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sdb.Close() })

	schemaSQL, err := os.ReadFile(filepath.Join("..", "..", "database", "schema.sqlite.sql"))
	if err != nil {
		t.Fatalf("read schema.sqlite.sql: %v", err)
	}
	if _, err := sdb.ExecContext(context.Background(), string(schemaSQL)); err != nil {
		t.Fatalf("apply schema.sqlite.sql: %v", err)
	}
	return sdb, db.NewForDriver(sdb, "sqlite")
}

func TestSameHostEmbeddedAndStandaloneWorkersClaimEachJobOnce(t *testing.T) {
	_, q := openSharedSQLiteFile(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := jobs.NewConfiguredService(q)
	for i := 1; i <= 6; i++ {
		if _, err := svc.Enqueue(ctx, jobs.EnqueueRequest{
			JobType:    jobs.JobTypeExternalLink,
			TargetType: jobs.JobTypeExternalLink,
			TargetKey:  "https://example.com/item-" + string(rune('0'+i)),
		}); err != nil {
			t.Fatalf("Enqueue %d: %v", i, err)
		}
	}

	var mu sync.Mutex
	seen := make(map[string]int)
	allDone := make(chan struct{})

	handler := func(_ context.Context, job *db.DurableJob) error {
		mu.Lock()
		seen[job.TargetKey]++
		count := len(seen)
		mu.Unlock()
		if count == 6 {
			select {
			case <-allDone:
			default:
				close(allDone)
			}
		}
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		svc.RunWorkerLoop(ctx, "embedded-worker", 10*time.Millisecond, []string{jobs.JobTypeExternalLink}, handler)
	}()
	go func() {
		defer wg.Done()
		standaloneSvc := jobs.NewConfiguredService(q)
		standaloneSvc.RunWorkerLoop(ctx, "standalone-worker", 10*time.Millisecond, []string{jobs.JobTypeExternalLink}, handler)
	}()

	select {
	case <-allDone:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for 6 jobs, got %d", len(seen))
	}
	cancel()
	wg.Wait()

	for k, v := range seen {
		if v != 1 {
			t.Fatalf("job %s executed %d times, want 1", k, v)
		}
	}
}

func TestDurableJobFailureNotifiesActiveAdminsUnread(t *testing.T) {
	sdb, q := openSharedSQLiteFile(t)
	ctx := context.Background()

	if _, err := sdb.ExecContext(ctx, `INSERT INTO roles (id, name, is_admin) VALUES (1, 'admin', 1), (2, 'member', 0)`); err != nil {
		t.Fatalf("insert roles: %v", err)
	}
	if _, err := sdb.ExecContext(ctx, `INSERT INTO users (idusers, username, deleted_at) VALUES (10, 'alice', NULL), (11, 'bob', NULL), (12, 'deleted_admin', CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("insert users: %v", err)
	}
	if _, err := sdb.ExecContext(ctx, `INSERT INTO user_roles (users_idusers, role_id) VALUES (10, 1), (11, 2), (12, 1)`); err != nil {
		t.Fatalf("insert user_roles: %v", err)
	}

	svc := jobs.NewConfiguredService(q)
	if _, err := svc.Enqueue(ctx, jobs.EnqueueRequest{
		JobType:     jobs.JobTypeExternalLink,
		TargetType:  jobs.JobTypeExternalLink,
		TargetKey:   "https://example.com/broken",
		MaxAttempts: 1,
	}); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	claimed, err := svc.ClaimNext(ctx, "worker-1", jobs.JobTypeExternalLink)
	if err != nil || claimed == nil {
		t.Fatalf("ClaimNext: (%v, %v)", claimed, err)
	}
	if applied, err := svc.FailOrRetryFenced(ctx, claimed, errors.New("connection refused")); err != nil || !applied {
		t.Fatalf("FailOrRetryFenced = (%v, %v)", applied, err)
	}

	aliceUnread, err := q.GetUnreadNotificationCountForLister(ctx, 10)
	if err != nil || aliceUnread != 1 {
		t.Fatalf("alice unread = (%d, %v), want (1, nil)", aliceUnread, err)
	}
	bobUnread, err := q.GetUnreadNotificationCountForLister(ctx, 11)
	if err != nil || bobUnread != 0 {
		t.Fatalf("bob unread = (%d, %v), want (0, nil)", bobUnread, err)
	}
	deletedUnread, err := q.GetUnreadNotificationCountForLister(ctx, 12)
	if err != nil || deletedUnread != 0 {
		t.Fatalf("deleted admin unread = (%d, %v), want (0, nil)", deletedUnread, err)
	}
}

func TestDurableJobFailureAggregatesUnreadByErrorClassAcrossHistoryPruning(t *testing.T) {
	database, q := openSharedSQLiteFile(t)
	ctx := t.Context()
	for _, statement := range []string{
		`INSERT INTO roles (id, name, is_admin) VALUES (1, 'admin', 1)`,
		`INSERT INTO users (idusers, username) VALUES (10, 'operator')`,
		`INSERT INTO user_roles (users_idusers, role_id) VALUES (10, 1)`,
	} {
		_, err := database.ExecContext(ctx, statement)
		require.NoError(t, err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	svc := jobs.NewConfiguredService(q)
	svc.Now = func() time.Time { return now }
	// target identifies the same URL through pruning and operator retries.
	const target = "https://example.test/private?key=fixture-secret"
	claimFailure := func() *jobs.ClaimedJob {
		t.Helper()
		created, err := svc.Enqueue(ctx, jobs.EnqueueRequest{
			JobType: jobs.JobTypeExternalLink, TargetType: jobs.JobTypeExternalLink, TargetKey: target,
			MaxAttempts: 1, ForceRetry: true,
		})
		require.NoError(t, err)
		require.True(t, created)
		claim, err := svc.ClaimNext(ctx, "worker", jobs.JobTypeExternalLink)
		require.NoError(t, err)
		require.NotNil(t, claim)
		return claim
	}
	fail := func(status int) {
		t.Helper()
		claim := claimFailure()
		applied, err := svc.FailOrRetryFenced(ctx, claim, &opengraph.HTTPStatusError{StatusCode: status})
		require.NoError(t, err)
		require.True(t, applied)
	}
	read := func() []*db.Notification {
		t.Helper()
		rows, err := q.ListUnreadNotificationsForLister(ctx, db.ListUnreadNotificationsForListerParams{ListerID: 10, Limit: 10})
		require.NoError(t, err)
		return rows
	}

	fail(http.StatusServiceUnavailable)
	original := read()
	require.Len(t, original, 1)
	pruned, err := svc.AdminPurge(ctx, jobs.StatusFailed)
	require.NoError(t, err)
	require.EqualValues(t, 1, pruned)
	now = now.Add(30 * time.Second)
	fail(http.StatusServiceUnavailable)
	repeated := read()
	require.Len(t, repeated, 1)
	require.Equal(t, original[0].ID, repeated[0].ID)
	require.Contains(t, repeated[0].Message.String, "2 failures")
	require.True(t, repeated[0].CreatedAt.After(original[0].CreatedAt))

	now = now.Add(30 * time.Second)
	fail(http.StatusNotFound)
	groups := read()
	require.Len(t, groups, 2, "a different failure class gets its own unread group")
	for _, group := range groups {
		require.NotContains(t, group.Message.String, "fixture-secret")
		require.NotContains(t, group.Link.String, "fixture-secret")
	}

	require.NoError(t, q.SetNotificationReadForLister(ctx, db.SetNotificationReadForListerParams{
		ID: repeated[0].ID, ListerID: 10,
	}))
	fail(http.StatusServiceUnavailable)
	require.Len(t, read(), 1, "the read group stays debounced")
	now = now.Add(svc.AdminNotifyDebounce + time.Second)
	fail(http.StatusServiceUnavailable)
	reopened, err := q.GetNotificationForLister(ctx, db.GetNotificationForListerParams{
		ID: repeated[0].ID, ListerID: 10,
	})
	require.NoError(t, err)
	require.False(t, reopened.ReadAt.Valid)
	require.Contains(t, reopened.Message.String, "4 failures")

	_, err = database.ExecContext(ctx, `CREATE TRIGGER reject_failure_summary BEFORE UPDATE OF message ON notifications
		WHEN NEW.resource_type = 'durable_job_failure' BEGIN SELECT RAISE(ABORT, 'summary unavailable'); END`)
	require.NoError(t, err)
	claim := claimFailure()
	applied, err := svc.FailOrRetryFenced(ctx, claim, &opengraph.HTTPStatusError{StatusCode: http.StatusBadGateway})
	require.ErrorContains(t, err, "summary unavailable")
	require.False(t, applied)
	active, err := q.SystemGetDurableJobByID(ctx, claim.Job.ID)
	require.NoError(t, err)
	require.Equal(t, jobs.StatusRunning, active.Status)
	require.Len(t, read(), 2, "no partial alert can survive a failed terminal transaction")
}

func TestSearchDirtyWorkerIndexesLatestRevision(t *testing.T) {
	sdb, q := openSharedSQLiteFile(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if _, err := sdb.ExecContext(ctx, `INSERT INTO users (idusers, username) VALUES (1, 'poster')`); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := sdb.ExecContext(ctx, `INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (42, 1, 1, 'initial words', CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("insert comment: %v", err)
	}

	svc := jobs.NewConfiguredService(q)
	if err := svc.EnqueueSearchDirty(ctx, searchworker.TypeComment, 42); err != nil {
		t.Fatalf("EnqueueSearchDirty 1: %v", err)
	}
	if _, err := sdb.ExecContext(ctx, `UPDATE comments SET text = 'updated durable searchable words' WHERE idcomments = 42`); err != nil {
		t.Fatalf("update comment: %v", err)
	}
	if err := svc.EnqueueSearchDirty(ctx, searchworker.TypeComment, 42); err != nil {
		t.Fatalf("EnqueueSearchDirty 2: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = config.RuntimeConfig{}
		searchworker.RunDurableWorker(ctx, svc, q, "search-test-worker")
	}()

	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		dirty, err := q.SystemGetSearchDirtyTarget(ctx, db.SystemGetSearchDirtyTargetParams{
			TargetType: searchworker.TypeComment,
			TargetID:   42,
		})
		if err == nil && dirty.Revision == 2 && dirty.IndexedRevision == 2 {
			cancel()
			wg.Wait()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	cancel()
	wg.Wait()
	t.Fatal("timed out waiting for search_dirty_targets indexed_revision to reach 2")
}

func TestThreadUpdateEnqueuesDurableWorkWithoutTaskEvent(t *testing.T) {
	_, q := openSharedSQLiteFile(t)
	ctx := t.Context()
	svc := jobs.NewConfiguredService(q)
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithJobService(svc))
	if err := cd.HandleThreadUpdated(ctx, common.ThreadUpdatedEvent{
		ThreadID:      7,
		CommentID:     42,
		CommentText:   "[url=https://example.com/reference]Reference[/url]",
		IncludeSearch: true,
	}); err != nil {
		t.Fatalf("apply thread side effects: %v", err)
	}

	for _, want := range []struct{ jobType, targetKey string }{
		{jobs.JobTypeSearchIndex, "comment:42"},
		{jobs.JobTypeExternalLink, "https://example.com/reference"},
	} {
		claimed, err := svc.ClaimNext(ctx, "worker", want.jobType)
		if err != nil || claimed == nil || claimed.Job.TargetKey != want.targetKey {
			t.Errorf("claim %s = (%v, %v), want target %q", want.jobType, claimed, err, want.targetKey)
		}
	}
}
