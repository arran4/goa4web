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
	migration, err := migrations.FS.ReadFile("0104_mysql.sql")
	require.NoError(t, err)
	up, _, ok := strings.Cut(string(migration), "-- +goose Down")
	require.True(t, ok)
	execReplySQL(t, conn, "CREATE TABLE schema_version (version INT NOT NULL)", up)
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
