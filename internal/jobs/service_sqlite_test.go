//go:build sqlite || sqlite3

package jobs

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/arran4/goa4web/internal/db"
)

func openTestJobsDB(t *testing.T) (*sql.DB, db.Querier) {
	t.Helper()
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	database.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = database.Close() })

	stmts := []string{
		`CREATE TABLE durable_jobs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			job_type TEXT NOT NULL,
			dedupe_key TEXT NOT NULL,
			active_dedupe_key TEXT DEFAULT NULL,
			target_type TEXT NOT NULL,
			target_id INTEGER NOT NULL DEFAULT 0,
			target_key TEXT NOT NULL DEFAULT '',
			target_revision INTEGER NOT NULL DEFAULT 0,
			needs_rerun INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'pending',
			attempts INTEGER NOT NULL DEFAULT 0,
			max_attempts INTEGER NOT NULL DEFAULT 5,
			run_after DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			lease_owner TEXT DEFAULT NULL,
			lease_token TEXT DEFAULT NULL,
			lease_expires_at DATETIME DEFAULT NULL,
			last_error TEXT DEFAULT NULL,
			created_by_user_id INTEGER DEFAULT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME DEFAULT NULL
		);`,
		`CREATE UNIQUE INDEX uq_durable_jobs_active_dedupe ON durable_jobs (active_dedupe_key);`,
		`CREATE TABLE durable_job_cooldowns (
			dedupe_key TEXT NOT NULL PRIMARY KEY,
			cooldown_until DATETIME NOT NULL,
			last_error TEXT DEFAULT NULL,
			last_notified_at DATETIME DEFAULT NULL,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE search_dirty_targets (
			target_type TEXT NOT NULL,
			target_id INTEGER NOT NULL,
			revision INTEGER NOT NULL DEFAULT 1,
			indexed_revision INTEGER NOT NULL DEFAULT 0,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (target_type, target_id)
		);`,
	}
	for _, s := range stmts {
		if _, err := database.Exec(s); err != nil {
			t.Fatalf("exec %q: %v", s, err)
		}
	}
	return database, db.NewSQLiteQuerier(database)
}

func TestDurableJobsEnqueueDedupAndNeedsRerun(t *testing.T) {
	_, q := openTestJobsDB(t)
	ctx := context.Background()
	now := time.Date(2026, time.November, 15, 12, 0, 0, 0, time.UTC)
	svc := NewService(q)
	svc.Now = func() time.Time { return now }

	created, err := svc.Enqueue(ctx, EnqueueRequest{
		JobType:    JobTypeExternalLink,
		TargetType: JobTypeExternalLink,
		TargetID:   11,
		TargetKey:  "https://example.com/a",
	})
	if err != nil || !created {
		t.Fatalf("first Enqueue = (%v, %v), want (true, nil)", created, err)
	}

	// Duplicate while pending coalesces cleanly.
	created2, err := svc.Enqueue(ctx, EnqueueRequest{
		JobType:    JobTypeExternalLink,
		TargetType: JobTypeExternalLink,
		TargetID:   11,
		TargetKey:  "https://example.com/a",
	})
	if err != nil || created2 {
		t.Fatalf("second Enqueue = (%v, %v), want (false, nil)", created2, err)
	}

	claimed, err := svc.ClaimNext(ctx, "worker-1", JobTypeExternalLink)
	if err != nil || claimed == nil {
		t.Fatalf("ClaimNext = (%v, %v)", claimed, err)
	}

	// Enqueue while running marks needs_rerun = true; completing re-arms one pending pass.
	if _, err := svc.Enqueue(ctx, EnqueueRequest{
		JobType:    JobTypeExternalLink,
		TargetType: JobTypeExternalLink,
		TargetID:   11,
		TargetKey:  "https://example.com/a",
	}); err != nil {
		t.Fatalf("Enqueue while running: %v", err)
	}

	applied, err := svc.CompleteFenced(ctx, claimed)
	if err != nil || !applied {
		t.Fatalf("CompleteFenced = (%v, %v), want (true, nil)", applied, err)
	}

	requeued, err := q.SystemGetDurableJobByID(ctx, claimed.Job.ID)
	if err != nil {
		t.Fatalf("reload job: %v", err)
	}
	if requeued.Status != StatusPending || requeued.NeedsRerun {
		t.Fatalf("after needs_rerun completion: status=%q needs_rerun=%v, want pending/false", requeued.Status, requeued.NeedsRerun)
	}

	// Claim and complete second pass -> terminal completed with active_dedupe_key cleared.
	claimed2, err := svc.ClaimNext(ctx, "worker-1", JobTypeExternalLink)
	if err != nil || claimed2 == nil {
		t.Fatalf("ClaimNext pass 2: %v", err)
	}
	if _, err := svc.CompleteFenced(ctx, claimed2); err != nil {
		t.Fatalf("CompleteFenced pass 2: %v", err)
	}
	done, err := q.SystemGetDurableJobByID(ctx, claimed.Job.ID)
	if err != nil {
		t.Fatalf("reload completed job: %v", err)
	}
	if done.Status != StatusCompleted || done.ActiveDedupeKey.Valid {
		t.Fatalf("completed job status=%q active_dedupe_key=%v", done.Status, done.ActiveDedupeKey)
	}
}

func TestDurableJobsLeaseExpiryFencingAndShutdownRelease(t *testing.T) {
	_, q := openTestJobsDB(t)
	ctx := context.Background()
	now := time.Date(2026, time.November, 15, 12, 0, 0, 0, time.UTC)
	svc := NewService(q)
	svc.Now = func() time.Time { return now }
	svc.LeaseDuration = 30 * time.Second

	if _, err := svc.Enqueue(ctx, EnqueueRequest{
		JobType:    JobTypeImageCache,
		TargetType: JobTypeImageCache,
		TargetKey:  "https://example.com/img.png",
	}); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	worker1Claim, err := svc.ClaimNext(ctx, "worker-1", JobTypeImageCache)
	if err != nil || worker1Claim == nil {
		t.Fatalf("worker1 ClaimNext: %v", err)
	}

	// Before lease expiry, worker-2 cannot claim it.
	if c, err := svc.ClaimNext(ctx, "worker-2", JobTypeImageCache); err != nil || c != nil {
		t.Fatalf("worker2 claimed before expiry: (%v, %v)", c, err)
	}

	// Advance past lease expiry -> worker-2 reclaims with a new fencing token.
	now = now.Add(31 * time.Second)
	worker2Claim, err := svc.ClaimNext(ctx, "worker-2", JobTypeImageCache)
	if err != nil || worker2Claim == nil {
		t.Fatalf("worker2 ClaimNext after expiry: (%v, %v)", worker2Claim, err)
	}
	if worker2Claim.LeaseToken == worker1Claim.LeaseToken {
		t.Fatal("expected distinct fencing token on reclaim")
	}

	// Stale worker-1 cannot complete or fail the job.
	if applied, err := svc.CompleteFenced(ctx, worker1Claim); err != nil || applied {
		t.Fatalf("stale CompleteFenced = (%v, %v), want (false, nil)", applied, err)
	}

	// Graceful shutdown release returns job to pending without charging worker-2's attempt.
	beforeAttempts := worker2Claim.Job.Attempts
	if applied, err := svc.ReleaseOnShutdownFenced(ctx, worker2Claim); err != nil || !applied {
		t.Fatalf("ReleaseOnShutdownFenced = (%v, %v)", applied, err)
	}
	after, err := q.SystemGetDurableJobByID(ctx, worker2Claim.Job.ID)
	if err != nil {
		t.Fatalf("reload job: %v", err)
	}
	if after.Status != StatusPending || after.Attempts != beforeAttempts-1 {
		t.Fatalf("after shutdown release: status=%q attempts=%d (before=%d)", after.Status, after.Attempts, beforeAttempts)
	}
}

func TestDurableJobsCooldownSurvivesPruningAndDebouncesAdminNotifications(t *testing.T) {
	_, q := openTestJobsDB(t)
	ctx := context.Background()
	now := time.Date(2026, time.November, 15, 12, 0, 0, 0, time.UTC)
	svc := NewService(q)
	svc.Now = func() time.Time { return now }
	svc.CooldownDuration = 24 * time.Hour
	svc.AdminNotifyDebounce = 1 * time.Hour

	notifications := 0
	svc.NotifyAdmins = func(_ context.Context, _, _ string) error {
		notifications++
		return nil
	}

	if _, err := svc.Enqueue(ctx, EnqueueRequest{
		JobType:     JobTypeExternalLink,
		TargetType:  JobTypeExternalLink,
		TargetKey:   "https://example.com/flaky",
		MaxAttempts: 1,
	}); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	claimed, err := svc.ClaimNext(ctx, "worker-1", JobTypeExternalLink)
	if err != nil || claimed == nil {
		t.Fatalf("ClaimNext: %v", err)
	}
	if applied, err := svc.FailOrRetryFenced(ctx, claimed, errors.New("upstream 503")); err != nil || !applied {
		t.Fatalf("FailOrRetryFenced: (%v, %v)", applied, err)
	}
	if notifications != 1 {
		t.Fatalf("notifications = %d, want 1", notifications)
	}

	// Prune terminal records immediately; cooldown must still block passive re-enqueue!
	now = now.Add(1 * time.Minute)
	pruned, err := svc.PruneTerminal(ctx, 0)
	if err != nil || pruned != 1 {
		t.Fatalf("PruneTerminal = (%d, %v), want (1, nil)", pruned, err)
	}
	if _, err := svc.Enqueue(ctx, EnqueueRequest{
		JobType:    JobTypeExternalLink,
		TargetType: JobTypeExternalLink,
		TargetKey:  "https://example.com/flaky",
	}); !errors.Is(err, ErrInCooldown) {
		t.Fatalf("Enqueue during cooldown after prune err = %v, want ErrInCooldown", err)
	}

	// Explicit ForceRetry bypasses cooldown; failing again within debounce window suppresses duplicate alert.
	if _, err := svc.Enqueue(ctx, EnqueueRequest{
		JobType:     JobTypeExternalLink,
		TargetType:  JobTypeExternalLink,
		TargetKey:   "https://example.com/flaky",
		MaxAttempts: 1,
		ForceRetry:  true,
	}); err != nil {
		t.Fatalf("ForceRetry Enqueue: %v", err)
	}
	claimed2, err := svc.ClaimNext(ctx, "worker-1", JobTypeExternalLink)
	if err != nil || claimed2 == nil {
		t.Fatalf("ClaimNext 2: %v", err)
	}
	if applied, err := svc.FailOrRetryFenced(ctx, claimed2, errors.New("still 503")); err != nil || !applied {
		t.Fatalf("FailOrRetryFenced 2: (%v, %v)", applied, err)
	}
	if notifications != 1 {
		t.Fatalf("debounced notifications = %d, want 1", notifications)
	}
}
