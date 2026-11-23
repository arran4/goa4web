//go:build sqlite

package jobs_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	require.NoError(t, searchworker.Index(ctx, q, searchworker.IndexEventData{
		Type: searchworker.TypeComment, ID: 42, Text: "initial words",
	}))

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

	marked, err := q.SystemMarkSearchTargetIndexed(ctx, db.SystemMarkSearchTargetIndexedParams{
		IndexedRevision: 1, TargetType: searchworker.TypeComment, TargetID: 42,
	})
	require.NoError(t, err)
	require.Zero(t, marked, "a superseded revision cannot be published")

	_, err = sdb.ExecContext(ctx, `CREATE TRIGGER reject_search_replacement BEFORE INSERT ON comments_search
		WHEN NEW.comment_id = 42 BEGIN SELECT RAISE(ABORT, 'search replacement unavailable'); END`)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		searchworker.RunDurableWorker(ctx, svc, q, "search-test-worker")
	}()

	defer func() {
		cancel()
		wg.Wait()
	}()
	waitIndexed := func(revision int64) {
		t.Helper()
		require.Eventually(t, func() bool {
			dirty, err := q.SystemGetSearchDirtyTarget(ctx, db.SystemGetSearchDirtyTargetParams{
				TargetType: searchworker.TypeComment, TargetID: 42,
			})
			return err == nil && dirty.Revision == revision && dirty.IndexedRevision == revision
		}, 4*time.Second, 20*time.Millisecond)
	}
	readWords := func() []string {
		t.Helper()
		rows, err := sdb.QueryContext(ctx, `SELECT word FROM searchwordlist sw
			JOIN comments_search sc ON sc.searchwordlist_idsearchwordlist = sw.idsearchwordlist
			WHERE sc.comment_id = 42`)
		require.NoError(t, err)
		defer func() { _ = rows.Close() }()
		var words []string
		for rows.Next() {
			var word string
			require.NoError(t, rows.Scan(&word))
			words = append(words, word)
		}
		require.NoError(t, rows.Err())
		return words
	}
	require.Eventually(t, func() bool {
		var status, lastError string
		err := sdb.QueryRowContext(ctx, `SELECT status, last_error FROM durable_jobs
			WHERE job_type = 'search_index' AND target_id = 42`).Scan(&status, &lastError)
		return err == nil && status == jobs.StatusPending && strings.Contains(lastError, "search replacement unavailable")
	}, 4*time.Second, 20*time.Millisecond)
	dirty, err := q.SystemGetSearchDirtyTarget(ctx, db.SystemGetSearchDirtyTargetParams{
		TargetType: searchworker.TypeComment, TargetID: 42,
	})
	require.NoError(t, err)
	require.Zero(t, dirty.IndexedRevision)
	require.ElementsMatch(t, []string{"initial", "words"}, readWords())

	_, err = sdb.ExecContext(ctx, `DROP TRIGGER reject_search_replacement`)
	require.NoError(t, err)
	_, err = svc.Enqueue(ctx, jobs.EnqueueRequest{
		JobType: jobs.JobTypeSearchIndex, TargetType: searchworker.TypeComment,
		TargetID: 42, TargetKey: "comment:42", TargetRevision: 2, ForceRetry: true,
	})
	require.NoError(t, err)
	waitIndexed(2)
	require.ElementsMatch(t, []string{"updated", "durable", "searchable", "words"}, readWords())

	_, err = sdb.ExecContext(ctx, `UPDATE comments SET deleted_at = CURRENT_TIMESTAMP WHERE idcomments = 42`)
	require.NoError(t, err)
	require.NoError(t, svc.EnqueueSearchDirty(ctx, searchworker.TypeComment, 42))
	waitIndexed(3)
	require.Empty(t, readWords())
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

func TestAdminEnqueueExternalURLCreatesAndReusesRecord(t *testing.T) {
	sdb, q := openSharedSQLiteFile(t)
	ctx := t.Context()
	now := time.Date(2026, time.November, 23, 12, 0, 0, 0, time.UTC)
	svc := jobs.NewConfiguredService(q)
	svc.Now = func() time.Time { return now }

	const target1 = "https://example.test/unused-one"
	linkID, created, err := svc.AdminEnqueueExternalURL(ctx, target1, sql.NullInt32{Int32: 7, Valid: true}, false)
	require.NoError(t, err)
	require.True(t, created)
	require.Positive(t, linkID)

	linkRow, err := q.GetExternalLinkByID(ctx, linkID)
	require.NoError(t, err)
	require.Equal(t, target1, linkRow.Url)

	listed, err := svc.AdminList(ctx, "", jobs.JobTypeExternalLink, 10, 0)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.Equal(t, int64(linkID), listed[0].TargetID)
	require.Equal(t, target1, listed[0].TargetKey)
	require.True(t, listed[0].CreatedByUserID.Valid)
	require.EqualValues(t, 7, listed[0].CreatedByUserID.Int32)

	// Advance SQLite last_insert_rowid() with a separate external_links row.
	_, err = sdb.ExecContext(ctx, `INSERT INTO external_links (url, clicks) VALUES ('https://example.test/other', 0)`)
	require.NoError(t, err)

	// Re-enqueueing while active coalesces and preserves the original link ID.
	reusedID, createdAgain, err := svc.AdminEnqueueExternalURL(ctx, target1, sql.NullInt32{Int32: 7, Valid: true}, false)
	require.NoError(t, err)
	require.False(t, createdAgain)
	require.Equal(t, linkID, reusedID)

	// Simulate cooldown after terminal failure and verify force bypasses it while reusing linkID.
	_, err = sdb.ExecContext(ctx, `UPDATE durable_jobs SET status = 'failed', active_dedupe_key = NULL WHERE id = ?`, listed[0].ID)
	require.NoError(t, err)
	_, err = sdb.ExecContext(ctx, `INSERT INTO durable_job_cooldowns (dedupe_key, cooldown_until, last_error) VALUES (?, ?, 'HTTP 503')`,
		jobs.DedupeKey(jobs.JobTypeExternalLink, target1), now.Add(time.Hour))
	require.NoError(t, err)

	_, _, err = svc.AdminEnqueueExternalURL(ctx, target1, sql.NullInt32{Int32: 7, Valid: true}, false)
	require.ErrorIs(t, err, jobs.ErrInCooldown)

	forcedID, createdForced, err := svc.AdminEnqueueExternalURL(ctx, target1, sql.NullInt32{Int32: 8, Valid: true}, true)
	require.NoError(t, err)
	require.True(t, createdForced)
	require.Equal(t, linkID, forcedID)

	var urlCount int
	require.NoError(t, sdb.QueryRowContext(ctx, `SELECT COUNT(*) FROM external_links WHERE url = ?`, target1).Scan(&urlCount))
	require.Equal(t, 1, urlCount)

	// Failed job insert must roll back creation of a brand-new external_links row.
	_, err = sdb.ExecContext(ctx, `CREATE TRIGGER reject_admin_link_job BEFORE INSERT ON durable_jobs
		WHEN NEW.target_key = 'https://example.test/rollback-unused'
		BEGIN SELECT RAISE(ABORT, 'injected admin job failure'); END`)
	require.NoError(t, err)
	_, _, err = svc.AdminEnqueueExternalURL(ctx, "https://example.test/rollback-unused", sql.NullInt32{Int32: 7, Valid: true}, false)
	require.ErrorContains(t, err, "injected admin job failure")

	require.NoError(t, sdb.QueryRowContext(ctx, `SELECT COUNT(*) FROM external_links WHERE url = 'https://example.test/rollback-unused'`).Scan(&urlCount))
	require.Zero(t, urlCount)
}

func TestAdminGetReturnsJobAndCooldownMetadata(t *testing.T) {
	sdb, q := openSharedSQLiteFile(t)
	ctx := t.Context()
	now := time.Date(2026, time.November, 23, 12, 0, 0, 0, time.UTC)
	svc := jobs.NewConfiguredService(q)
	svc.Now = func() time.Time { return now }

	_, err := svc.AdminGet(ctx, 999999)
	require.ErrorIs(t, err, sql.ErrNoRows)

	// target is an external URL exercised by the AdminGet metadata test.
	const target = "https://example.test/show-meta"
	linkID, created, err := svc.AdminEnqueueExternalURL(ctx, target, sql.NullInt32{Int32: 4, Valid: true}, false)
	require.NoError(t, err)
	require.True(t, created)

	list, err := svc.AdminList(ctx, "", "", 10, 0)
	require.NoError(t, err)
	require.Len(t, list, 1)

	details, err := svc.AdminGet(ctx, list[0].ID)
	require.NoError(t, err)
	require.Equal(t, int64(linkID), details.Job.TargetID)
	require.Nil(t, details.Cooldown)

	cooldownUntil := now.Add(2 * time.Hour)
	_, err = sdb.ExecContext(ctx, `INSERT INTO durable_job_cooldowns (dedupe_key, cooldown_until, last_error) VALUES (?, ?, 'HTTP 502')`,
		details.Job.DedupeKey, cooldownUntil)
	require.NoError(t, err)

	details, err = svc.AdminGet(ctx, list[0].ID)
	require.NoError(t, err)
	require.NotNil(t, details.Cooldown)
	require.True(t, details.Cooldown.CooldownUntil.Equal(cooldownUntil))
}

func TestWorkerShutdownAtHandlerReturnNeverStrandsRunningLease(t *testing.T) {
	_, q := openSharedSQLiteFile(t)
	ctx := t.Context()
	svc := jobs.NewConfiguredService(q)

	for i := 0; i < 15; i++ {
		targetID := int64(i + 1)
		created, err := svc.Enqueue(ctx, jobs.EnqueueRequest{
			JobType:    jobs.JobTypeSearchIndex,
			TargetType: searchworker.TypeComment,
			TargetID:   targetID,
			TargetKey:  fmt.Sprintf("comment:%d", targetID),
		})
		require.NoError(t, err)
		require.True(t, created)

		pending, err := svc.AdminList(ctx, jobs.StatusPending, jobs.JobTypeSearchIndex, 10, 0)
		require.NoError(t, err)
		require.Len(t, pending, 1)
		jobID := pending[0].ID

		wCtx, cancel := context.WithCancel(ctx)
		returning := make(chan struct{}, 1)
		done := make(chan struct{})
		go func() {
			defer close(done)
			svc.RunWorkerLoop(wCtx, "shutdown-race-worker", 5*time.Millisecond, []string{jobs.JobTypeSearchIndex}, func(context.Context, *db.DurableJob) error {
				returning <- struct{}{}
				return nil
			})
		}()

		<-returning
		cancel()
		<-done

		details, err := svc.AdminGet(ctx, jobID)
		require.NoError(t, err)
		require.NotEqual(t, jobs.StatusRunning, details.Job.Status)
		require.Contains(t, []string{jobs.StatusCompleted, jobs.StatusPending}, details.Job.Status)
		require.False(t, details.Job.LeaseToken.Valid)
		if details.Job.Status == jobs.StatusPending {
			cancelled, err := svc.AdminCancel(ctx, jobID)
			require.NoError(t, err)
			require.True(t, cancelled)
		}
	}
}
