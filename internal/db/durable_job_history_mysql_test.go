//go:build !sqlite && !sqlite3

package db

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/migrations"
	"github.com/stretchr/testify/require"
)

func TestMySQLPruneTerminalDurableJobHistory(t *testing.T) {
	ctx := t.Context()
	conn := openReplyThreadTestDatabase(t)
	loadMySQLDurableJobMigration(t, conn)
	execReplySQL(t, conn, `
		INSERT INTO durable_jobs (id, job_type, dedupe_key, target_type, status, completed_at)
		VALUES (1, 'external_link', 'first', 'external_link', 'failed', '2026-11-15 11:00:00'),
		       (2, 'external_link', 'second', 'external_link', 'completed', '2026-11-15 11:00:00'),
		       (3, 'image_cache', 'old', 'image_cache', 'cancelled', '2026-11-14 11:00:00'),
		       (4, 'image_cache', 'held', 'image_cache', 'running', NULL),
		       (5, 'image_cache', 'recent', 'image_cache', 'failed', '2026-11-15 11:00:00')`)
	q := NewForDriver(conn, "mysql")
	n, err := q.SystemPruneTerminalDurableJobs(ctx, SystemPruneTerminalDurableJobsParams{
		Cutoff:     sql.NullTime{Time: time.Date(2026, time.November, 14, 12, 0, 0, 0, time.UTC), Valid: true},
		MaxPerKind: 1,
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), n)
	for _, id := range []int64{2, 4, 5} {
		_, err := q.SystemGetDurableJobByID(ctx, id)
		require.NoError(t, err, "job %d should remain", id)
	}
}

func loadMySQLDurableJobMigration(t *testing.T, conn *sql.DB) {
	t.Helper()
	migration, err := migrations.FS.ReadFile("0104_mysql.sql")
	require.NoError(t, err)
	up, _, ok := strings.Cut(string(migration), "-- +goose Down")
	require.True(t, ok)
	execReplySQL(t, conn, "CREATE TABLE schema_version (version INT NOT NULL)", up)
}

func TestMySQLQueuePendingImageCacheJobs(t *testing.T) {
	ctx := t.Context()
	conn := openReplyThreadTestDatabase(t)
	loadMySQLDurableJobMigration(t, conn)
	imageMigration, err := migrations.FS.ReadFile("0085_mysql.sql")
	require.NoError(t, err)
	execReplySQL(t, conn, string(imageMigration))
	_, err = conn.ExecContext(ctx, `
		INSERT INTO image_cache_entries (id, source_url, source_kind, status, next_attempt_at)
		VALUES ('recover.png', ?, 'remote', 'pending', NULL),
		       ('active.png', 'https://example.test/active.png', 'remote', 'pending', NULL),
		       ('cooldown.png', 'https://example.test/cooldown.png', 'remote', 'pending', NULL),
		       ('future.png', 'https://example.test/future.png', 'remote', 'pending', '2030-01-01'),
		       ('ready.png', 'https://example.test/ready.png', 'remote', 'ready', NULL),
		       ('failed.png', 'https://example.test/failed.png', 'remote', 'failed', NULL)`,
		"https://example.test/recover.png?"+strings.Repeat("q", 600))
	require.NoError(t, err)
	execReplySQL(t, conn,
		`INSERT INTO durable_jobs (job_type, dedupe_key, active_dedupe_key, target_type, target_key)
		  VALUES ('image_cache', 'image_cache:https://example.test/active.png',
		          'image_cache:https://example.test/active.png', 'image_cache', 'https://example.test/active.png')`,
		`INSERT INTO durable_job_cooldowns (dedupe_key, cooldown_until)
		  VALUES ('image_cache_entry:cooldown.png', '2030-01-01')`,
	)
	q := NewForDriver(conn, "mysql")
	now := time.Date(2026, time.November, 15, 12, 0, 0, 0, time.UTC)
	params := SystemQueuePendingImageCacheJobsParams{
		MaxAttempts: 4, RunAfter: now, DueAt: sql.NullTime{Time: now, Valid: true}, RetryFailed: false, CooldownAt: now,
	}
	n, err := q.SystemQueuePendingImageCacheJobs(ctx, params)
	require.NoError(t, err)
	require.Equal(t, int64(1), n)

	n, err = q.SystemQueuePendingImageCacheJobs(ctx, params)
	require.NoError(t, err)
	require.Zero(t, n)

	params.EntryID = sql.NullString{String: "failed.png", Valid: true}
	params.RetryFailed = true
	params.ActorID = sql.NullInt32{Int32: 42, Valid: true}
	n, err = q.SystemQueuePendingImageCacheJobs(ctx, params)
	require.NoError(t, err)
	require.Equal(t, int64(1), n)
	rows, err := q.AdminListDurableJobs(ctx, AdminListDurableJobsParams{Column1: "", Column3: "", Limit: 10})
	require.NoError(t, err)
	require.Len(t, rows, 3)
	require.Equal(t, "failed.png", rows[0].TargetKey)
	require.Equal(t, params.ActorID, rows[0].CreatedByUserID)
	require.Equal(t, "recover.png", rows[1].TargetKey)
	require.Equal(t, "image_cache_entry", rows[1].TargetType)
}
