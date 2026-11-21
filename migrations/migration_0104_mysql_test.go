package migrations

import (
	"fmt"
	"os"
	"testing"

	"github.com/arran4/goa4web/database"
	"github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
)

func TestMySQLMigration0104UpgradePathsAndDown(t *testing.T) {
	dsn := os.Getenv("GOA4WEB_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set GOA4WEB_TEST_MYSQL_DSN to run MySQL/MariaDB migration tests")
	}
	config, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("parse MySQL DSN: %v", err)
	}
	config.MultiStatements = true

	for _, start := range []int64{101, 102, 103} {
		t.Run(fmt.Sprintf("from_%d", start), func(t *testing.T) {
			ctx := t.Context()
			conn := openTemporaryMySQLDatabase(t, config.FormatDSN())
			// Bootstrap the same way as a fresh server. Ancient MySQL migrations
			// predate Goose and are not the supported empty-database initializer.
			if _, err := conn.ExecContext(ctx, string(database.SchemaMySQL)); err != nil {
				t.Fatalf("load canonical MySQL schema: %v", err)
			}
			if _, err := conn.ExecContext(ctx, "CREATE TABLE schema_version (version INT NOT NULL); INSERT INTO schema_version VALUES (104)"); err != nil {
				t.Fatalf("initialize legacy schema version: %v", err)
			}
			// Canonical snapshots list recent migrations. Model the complete
			// ledger of an installation transitioned from the legacy version.
			for version := 1; version < 92; version++ {
				if _, err := conn.ExecContext(ctx, "INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, 1)", version); err != nil {
					t.Fatalf("record historical migration %d: %v", version, err)
				}
			}
			provider, err := goose.NewProvider(goose.DialectMySQL, conn, ForDriver("mysql"))
			if err != nil {
				t.Fatalf("create MySQL migration provider: %v", err)
			}
			if _, err := provider.DownTo(ctx, start); err != nil {
				t.Fatalf("prepare empty version %d schema: %v", start, err)
			}
			if _, err := conn.ExecContext(ctx, "INSERT INTO notifications (id, users_idusers, message) VALUES (42, 1, 'existing notification')"); err != nil {
				t.Fatalf("insert pre-upgrade notification: %v", err)
			}
			assertVersion := func(want int64) {
				t.Helper()
				var legacy int64
				if err := conn.QueryRowContext(ctx, "SELECT version FROM schema_version").Scan(&legacy); err != nil || legacy != want {
					t.Fatalf("legacy version = %d, err %v, want %d", legacy, err, want)
				}
				if got, err := provider.GetDBVersion(ctx); err != nil || got != want {
					t.Fatalf("Goose version = %d, err %v, want %d", got, err, want)
				}
			}
			assertVersion(start)
			if _, err := provider.UpTo(ctx, 103); err != nil {
				t.Fatalf("upgrade from %d to 103: %v", start, err)
			}
			if _, err := conn.ExecContext(ctx, "UPDATE notifications SET activity_key = 'forum.reply:7:11', resource_id = 7, activity_position = 11 WHERE id = 42"); err != nil {
				t.Fatalf("update notification activity: %v", err)
			}
			if _, err := provider.UpTo(ctx, 104); err != nil {
				t.Fatalf("upgrade from 103 to 104: %v", err)
			}
			assertVersion(104)
			if _, err := conn.ExecContext(ctx, "INSERT INTO durable_jobs (job_type, dedupe_key, active_dedupe_key, target_type) VALUES ('search.index', 'forum_thread:7', 'forum_thread:7', 'forum_thread')"); err != nil {
				t.Fatalf("insert durable job: %v", err)
			}
			if _, err := provider.DownTo(ctx, 103); err != nil {
				t.Fatalf("downgrade from 104 to 103: %v", err)
			}
			assertVersion(103)
			var message, activityKey string
			var resourceID, position int64
			if err := conn.QueryRowContext(ctx, "SELECT message, activity_key, resource_id, activity_position FROM notifications WHERE id = 42").Scan(&message, &activityKey, &resourceID, &position); err != nil {
				t.Fatalf("read notification after downgrade: %v", err)
			}
			if message != "existing notification" || activityKey != "forum.reply:7:11" || resourceID != 7 || position != 11 {
				t.Fatalf("notification changed during migration: %q, %q, %d, %d", message, activityKey, resourceID, position)
			}
			var tableCount int
			if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name IN ('durable_jobs', 'durable_job_cooldowns', 'search_dirty_targets')").Scan(&tableCount); err != nil || tableCount != 0 {
				t.Fatalf("durable job tables after downgrade = %d, err %v, want 0", tableCount, err)
			}
		})
	}
}
