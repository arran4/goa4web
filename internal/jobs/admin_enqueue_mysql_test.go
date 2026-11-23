//go:build !sqlite && !sqlite3

package jobs_test

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/searchindex"
	"github.com/arran4/goa4web/migrations"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func openMySQLJobsTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("GOA4WEB_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set GOA4WEB_TEST_MYSQL_DSN to run MySQL/MariaDB durable job tests")
	}
	cfg, err := mysql.ParseDSN(dsn)
	require.NoError(t, err)
	dbName := fmt.Sprintf("goa4web_jobs_%d", time.Now().UnixNano())
	adminCfg := *cfg
	adminCfg.DBName = ""
	adminDB, err := sql.Open("mysql", adminCfg.FormatDSN())
	require.NoError(t, err)
	_, err = adminDB.Exec("CREATE DATABASE `" + dbName + "`")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = adminDB.Exec("DROP DATABASE `" + dbName + "`")
		_ = adminDB.Close()
	})
	testCfg := *cfg
	testCfg.DBName = dbName
	testCfg.MultiStatements = true
	conn, err := sql.Open("mysql", testCfg.FormatDSN())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	migration, err := migrations.FS.ReadFile("0104_mysql.sql")
	require.NoError(t, err)
	up, _, ok := strings.Cut(string(migration), "-- +goose Down")
	require.True(t, ok)
	up = strings.ReplaceAll(up, "-- +goose Up", "")
	up = strings.ReplaceAll(up, "-- +goose StatementBegin", "")
	up = strings.ReplaceAll(up, "-- +goose StatementEnd", "")

	_, err = conn.ExecContext(t.Context(), `
		CREATE TABLE IF NOT EXISTS schema_version (version INT NOT NULL);
		DELETE FROM schema_version;
		INSERT INTO schema_version VALUES (103);
		CREATE TABLE external_links (
			id int NOT NULL AUTO_INCREMENT,
			url text NOT NULL,
			url_hash binary(32) GENERATED ALWAYS AS (unhex(sha2(url, 256))) STORED,
			clicks int NOT NULL DEFAULT 0,
			created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			updated_by int DEFAULT NULL,
			card_title tinytext,
			card_description text,
			card_image text,
			card_image_cache varchar(255) DEFAULT NULL,
			favicon_cache varchar(255) DEFAULT NULL,
			card_duration tinytext,
			card_upload_date tinytext,
			card_author tinytext,
			PRIMARY KEY (id),
			UNIQUE KEY external_links_url_hash_idx (url_hash)
		);
	`+up)
	require.NoError(t, err)
	return conn
}

func TestMySQLAdminEnqueueExternalURLAndAdminGet(t *testing.T) {
	conn := openMySQLJobsTestDB(t)
	ctx := t.Context()
	q := db.NewForDriver(conn, "mysql")
	now := time.Date(2026, time.November, 23, 12, 0, 0, 0, time.UTC)
	svc := jobs.NewConfiguredService(q)
	svc.Now = func() time.Time { return now }

	// target is an external URL exercised by the live MySQL admin enqueue test.
	const target = "https://example.test/mysql-unused"
	linkID, created, err := svc.AdminEnqueueExternalURL(ctx, target, sql.NullInt32{Int32: 5, Valid: true}, false)
	require.NoError(t, err)
	require.True(t, created)
	require.Positive(t, linkID)

	list, err := svc.AdminList(ctx, "", jobs.JobTypeExternalLink, 10, 0)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, int64(linkID), list[0].TargetID)

	_, err = svc.AdminGet(ctx, 999999)
	require.ErrorIs(t, err, sql.ErrNoRows)

	details, err := svc.AdminGet(ctx, list[0].ID)
	require.NoError(t, err)
	require.Equal(t, int64(linkID), details.Job.TargetID)
	require.Nil(t, details.Cooldown)

	// Re-enqueue while active reuses the exact external_links row.
	_, err = conn.ExecContext(ctx, `INSERT INTO external_links (url, clicks) VALUES ('https://example.test/mysql-other', 0)`)
	require.NoError(t, err)
	reusedID, createdAgain, err := svc.AdminEnqueueExternalURL(ctx, target, sql.NullInt32{Int32: 5, Valid: true}, false)
	require.NoError(t, err)
	require.False(t, createdAgain)
	require.Equal(t, linkID, reusedID)

	// Simulate terminal cooldown and confirm force=true bypasses cooldown and reuses linkID.
	_, err = conn.ExecContext(ctx, `UPDATE durable_jobs SET status = 'failed', active_dedupe_key = NULL WHERE id = ?`, list[0].ID)
	require.NoError(t, err)
	cooldownUntil := now.Add(time.Hour)
	_, err = conn.ExecContext(ctx, `INSERT INTO durable_job_cooldowns (dedupe_key, cooldown_until, last_error) VALUES (?, ?, 'HTTP 503')`,
		list[0].DedupeKey, cooldownUntil)
	require.NoError(t, err)

	details, err = svc.AdminGet(ctx, list[0].ID)
	require.NoError(t, err)
	require.NotNil(t, details.Cooldown)

	_, _, err = svc.AdminEnqueueExternalURL(ctx, target, sql.NullInt32{Int32: 5, Valid: true}, false)
	require.ErrorIs(t, err, jobs.ErrInCooldown)

	forcedID, createdForced, err := svc.AdminEnqueueExternalURL(ctx, target, sql.NullInt32{Int32: 5, Valid: true}, true)
	require.NoError(t, err)
	require.True(t, createdForced)
	require.Equal(t, linkID, forcedID)

	var urlCount int
	require.NoError(t, conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM external_links WHERE url = ?`, target).Scan(&urlCount))
	require.Equal(t, 1, urlCount)
}

func TestMySQLSameSecondCoalesceSearchAndClaimComplete(t *testing.T) {
	conn := openMySQLJobsTestDB(t)
	ctx := t.Context()
	q := db.NewForDriver(conn, "mysql")
	now := time.Date(2026, time.November, 23, 12, 5, 0, 0, time.UTC)
	svc := jobs.NewConfiguredService(q)
	svc.Now = func() time.Time { return now }

	// Two rapid search-dirty bumps in the same second coalesce to one pending job at revision 2.
	require.NoError(t, svc.EnqueueSearchDirty(ctx, searchindex.TypeComment, 77))
	require.NoError(t, svc.EnqueueSearchDirty(ctx, searchindex.TypeComment, 77))

	pending, err := svc.AdminList(ctx, jobs.StatusPending, jobs.JobTypeSearchIndex, 10, 0)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.Equal(t, int64(2), pending[0].TargetRevision)

	claimed, err := svc.ClaimNext(ctx, "mysql-worker", jobs.JobTypeSearchIndex)
	require.NoError(t, err)
	require.NotNil(t, claimed)
	require.Equal(t, int64(2), claimed.Job.TargetRevision)

	// Coalesce while running (including identical same-second Enqueue calls that change zero columns).
	require.NoError(t, svc.EnqueueSearchDirty(ctx, searchindex.TypeComment, 77))
	require.NoError(t, svc.EnqueueSearchDirty(ctx, searchindex.TypeComment, 77))
	created, err := svc.Enqueue(ctx, jobs.EnqueueRequest{
		JobType:    jobs.JobTypeSearchIndex,
		TargetType: searchindex.TypeComment,
		TargetID:   77,
		TargetKey:  "comment:77",
	})
	require.NoError(t, err)
	require.False(t, created)

	ok, err := svc.CompleteFenced(ctx, claimed)
	require.NoError(t, err)
	require.True(t, ok)

	rearmed, err := svc.AdminGet(ctx, claimed.Job.ID)
	require.NoError(t, err)
	require.Equal(t, jobs.StatusPending, rearmed.Job.Status)
	require.False(t, rearmed.Job.NeedsRerun)
	require.Equal(t, int64(4), rearmed.Job.TargetRevision)

	claimed2, err := svc.ClaimNext(ctx, "mysql-worker", jobs.JobTypeSearchIndex)
	require.NoError(t, err)
	require.NotNil(t, claimed2)
	ok, err = svc.CompleteFenced(ctx, claimed2)
	require.NoError(t, err)
	require.True(t, ok)

	done, err := svc.AdminGet(ctx, claimed.Job.ID)
	require.NoError(t, err)
	require.Equal(t, jobs.StatusCompleted, done.Job.Status)
	require.False(t, done.Job.NeedsRerun)
	require.Equal(t, int64(4), done.Job.TargetRevision)
}
