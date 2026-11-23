package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/app"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/internal/sqlutil"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/schema_v98_mysql.sql
var schemaV98MySQL string

type mysqlUpgradeHarness struct {
	appDSN       string
	migrationDSN string
	appConn      *sql.DB
	migConn      *sql.DB
}

func newMySQLUpgradeHarness(t *testing.T, baseDSN, prefix, charset string) *mysqlUpgradeHarness {
	t.Helper()
	ctx := t.Context()
	cfg, err := mysql.ParseDSN(baseDSN)
	require.NoError(t, err)
	cfg.DBName = ""
	admin, err := sql.Open("mysql", cfg.FormatDSN())
	require.NoError(t, err)
	t.Cleanup(func() { _ = admin.Close() })

	name := fmt.Sprintf("goa4web_%s_%d", prefix, time.Now().UnixNano())
	_, err = admin.ExecContext(ctx, "CREATE DATABASE `"+name+"` DEFAULT CHARACTER SET "+charset)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = admin.Exec("DROP DATABASE `" + name + "`") })

	cfg.DBName = name
	cfg.MultiStatements = false
	appDSN := cfg.FormatDSN()
	cfg.MultiStatements = true
	migrationDSN := cfg.FormatDSN()

	appConn, err := sql.Open("mysql", appDSN)
	require.NoError(t, err)
	t.Cleanup(func() { _ = appConn.Close() })

	migConn, err := sql.Open("mysql", migrationDSN)
	require.NoError(t, err)
	t.Cleanup(func() { _ = migConn.Close() })

	return &mysqlUpgradeHarness{
		appDSN:       appDSN,
		migrationDSN: migrationDSN,
		appConn:      appConn,
		migConn:      migConn,
	}
}

func (h *mysqlUpgradeHarness) runMigrate(t *testing.T, withMigrationConn bool) error {
	t.Helper()
	args := []string{"goa4web", "--db-driver", "mysql", "--db-conn", h.appDSN}
	if withMigrationConn {
		args = append(args, "--db-migration-conn", h.migrationDSN)
	}
	args = append(args, "db", "migrate")
	root, err := parseRoot(args)
	if err != nil {
		return err
	}
	defer root.Close()
	return root.Run()
}

func (h *mysqlUpgradeHarness) loadV98Schema(t *testing.T, withGooseHistory bool) {
	t.Helper()
	ctx := t.Context()
	require.NoError(t, sqlutil.RunStatements(ctx, h.appConn, strings.NewReader(schemaV98MySQL)))
	_, err := h.appConn.ExecContext(ctx, "INSERT INTO schema_version (version) VALUES (98)")
	require.NoError(t, err)
	if !withGooseHistory {
		_, err := h.appConn.ExecContext(ctx, "DROP TABLE IF EXISTS goose_db_version")
		require.NoError(t, err)
		return
	}
	for version := 0; version <= 98; version++ {
		_, err := h.appConn.ExecContext(ctx, "INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, 1)", version)
		require.NoError(t, err)
	}
}

func TestProductionMySQLUpgradeFromV98AndStartupSmoke(t *testing.T) {
	dsn := os.Getenv("GOA4WEB_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set GOA4WEB_TEST_MYSQL_DSN to run MySQL/MariaDB upgrade tests")
	}

	for _, tc := range []struct {
		name             string
		withGooseHistory bool
	}{
		{name: "v98_legacy_transition", withGooseHistory: false},
		{name: "v98_goose_history", withGooseHistory: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			h := newMySQLUpgradeHarness(t, dsn, tc.name, "latin1")
			h.loadV98Schema(t, tc.withGooseHistory)

			// Seed representative, user-data-free pre-upgrade rows across affected tables.
			userA := strings.Repeat("u", 254) + "a"
			userB := strings.Repeat("u", 254) + "b"
			_, err := h.appConn.ExecContext(ctx, "INSERT INTO roles (id, name, can_login, is_admin) VALUES (1, 'user', 1, 0), (2, 'administrator', 1, 1)")
			require.NoError(t, err)
			_, err = h.appConn.ExecContext(ctx, "INSERT INTO users (idusers, username) VALUES (1, ?), (2, ?)", userA, userB)
			require.NoError(t, err)
			_, err = h.appConn.ExecContext(ctx, "INSERT INTO user_emails (user_id, email) VALUES (1, 'fixture@example.invalid')")
			require.NoError(t, err)
			_, err = h.appConn.ExecContext(ctx, "INSERT INTO preferences (users_idusers, emailforumupdates, daily_digest_hour) VALUES (1, 1, 8)")
			require.NoError(t, err)
			_, err = h.appConn.ExecContext(ctx, "INSERT INTO notifications (id, users_idusers, message) VALUES (42, 1, 'pre-upgrade notice')")
			require.NoError(t, err)
			_, err = h.appConn.ExecContext(ctx, "INSERT INTO pending_emails (id, to_user_id, body) VALUES (34, 1, 'pre-upgrade mime')")
			require.NoError(t, err)
			_, err = h.appConn.ExecContext(ctx, "INSERT INTO external_links (url, clicks) VALUES ('https://example.invalid/v98', 5)")
			require.NoError(t, err)
			_, err = h.appConn.ExecContext(ctx, "INSERT INTO searchwordlist (word) VALUES ('legacyword')")
			require.NoError(t, err)
			_, err = h.appConn.ExecContext(ctx, "INSERT INTO content_read_markers (item, item_id, user_id, last_comment_id) VALUES ('forum_thread', 7, 1, 11)")
			require.NoError(t, err)

			// Mismatched DSN guard: pending migration must reject multiStatements=false.
			err = h.runMigrate(t, false)
			require.Error(t, err)
			require.Contains(t, err.Error(), "DB_MIGRATION_CONN")
			var version int
			require.NoError(t, h.appConn.QueryRowContext(ctx, "SELECT version FROM schema_version").Scan(&version))
			require.Equal(t, 98, version)

			// Upgrade through real Goose startup path using separate migration DSN.
			require.NoError(t, h.runMigrate(t, true))
			// Idempotent second run works even with only appDSN (multiStatements=false).
			require.NoError(t, h.runMigrate(t, false))

			require.NoError(t, h.appConn.QueryRowContext(ctx, "SELECT version FROM schema_version").Scan(&version))
			require.Equal(t, handlers.ExpectedSchemaVersion, version)

			var count, minVer, maxVer int
			require.NoError(t, h.appConn.QueryRowContext(ctx, "SELECT COUNT(DISTINCT version_id), MIN(version_id), MAX(version_id) FROM goose_db_version WHERE is_applied = 1").Scan(&count, &minVer, &maxVer))
			require.Equal(t, 0, minVer)
			require.Equal(t, handlers.ExpectedSchemaVersion, maxVer)
			require.Equal(t, handlers.ExpectedSchemaVersion+1, count)

			// Verify pre-upgrade data and column semantics survived migrations 0099..0107.
			var gotA, gotB, gotEmail, gotNotice, gotBody, gotWord string
			var gotClicks int
			var gotGen, gotHour int
			var gotEnabled bool
			var gotPos int64
			require.NoError(t, h.appConn.QueryRowContext(ctx, "SELECT username FROM users WHERE idusers = 1").Scan(&gotA))
			require.NoError(t, h.appConn.QueryRowContext(ctx, "SELECT username FROM users WHERE idusers = 2").Scan(&gotB))
			require.Equal(t, userA, gotA)
			require.Equal(t, userB, gotB)
			require.NoError(t, h.appConn.QueryRowContext(ctx, "SELECT email FROM user_emails WHERE user_id = 1").Scan(&gotEmail))
			require.Equal(t, "fixture@example.invalid", gotEmail)
			require.NoError(t, h.appConn.QueryRowContext(ctx, "SELECT optional_email_generation, emailforumupdates, daily_digest_hour FROM preferences WHERE users_idusers = 1").Scan(&gotGen, &gotEnabled, &gotHour))
			require.Zero(t, gotGen)
			require.True(t, gotEnabled)
			require.Equal(t, 8, gotHour)
			require.NoError(t, h.appConn.QueryRowContext(ctx, "SELECT message FROM notifications WHERE id = 42").Scan(&gotNotice))
			require.Equal(t, "pre-upgrade notice", gotNotice)
			require.NoError(t, h.appConn.QueryRowContext(ctx, "SELECT body FROM pending_emails WHERE id = 34").Scan(&gotBody))
			require.Equal(t, "pre-upgrade mime", gotBody)
			require.NoError(t, h.appConn.QueryRowContext(ctx, "SELECT clicks FROM external_links WHERE url_hash = UNHEX(SHA2('https://example.invalid/v98', 256))").Scan(&gotClicks))
			require.Equal(t, 5, gotClicks)
			require.NoError(t, h.appConn.QueryRowContext(ctx, "SELECT word FROM searchwordlist WHERE word = 'legacyword'").Scan(&gotWord))
			require.Equal(t, "legacyword", gotWord)
			require.NoError(t, h.appConn.QueryRowContext(ctx, "SELECT last_activity_position FROM content_read_markers WHERE item='forum_thread' AND item_id=7 AND user_id=1").Scan(&gotPos))
			require.Equal(t, int64(11)<<32, gotPos)

			// Exercise minimal representative inserts into migrated structures using appDSN.
			_, err = h.appConn.ExecContext(ctx, "INSERT INTO pending_actions (id, uid, browser_id, action_type, form_data, expires_at) VALUES ('pa1', 1, 'b1', 'test', '{}', DATE_ADD(CURRENT_TIMESTAMP, INTERVAL 1 HOUR))")
			require.NoError(t, err)
			_, err = h.appConn.ExecContext(ctx, "INSERT INTO notifications (users_idusers, message, activity_key, resource_type, resource_id, activity_position) VALUES (1, 'new activity', 'forum.reply:7:12', 'forum_thread', 7, 12)")
			require.NoError(t, err)
			_, err = h.appConn.ExecContext(ctx, "INSERT INTO durable_jobs (job_type, dedupe_key, active_dedupe_key, target_type, target_id, target_key) VALUES ('search_index', 'search_index:comment:1', 'search_index:comment:1', 'comment', 1, 'comment:1')")
			require.NoError(t, err)
			_, err = h.appConn.ExecContext(ctx, "INSERT INTO forum_reply_email_activities (recipient_id, activity_position, thread_id, window_end) VALUES (1, 12, 7, 999999999)")
			require.NoError(t, err)

			// Startup checks + server HTTP readiness smoke with AutoMigrate=true and multiStatements=false.
			root, err := parseRoot([]string{"goa4web", "--db-driver", "mysql", "--db-conn", h.appDSN})
			require.NoError(t, err)
			defer root.Close()

			cfg := *root.cfg
			cfg.DBMigrationConn = ""
			cfg.AutoMigrate = true
			cfg.ImageUploadDir = t.TempDir()
			cfg.ImageCacheDir = t.TempDir()
			cfg.CreateDirs = true
			cfg.BaseURL = "http://127.0.0.1"

			checkedPool, err := dbstart.PerformStartupChecks(&cfg, root.dbReg)
			require.NoError(t, err)
			require.NoError(t, checkedPool.Close())

			srv, err := app.NewServer(ctx, &cfg, root.adminHandlers,
				app.WithSessionSecret("test-session-secret-0123456789abcdef"),
				app.WithImageSignSecret("test-image-sign-secret-0123456789"),
				app.WithLinkSignSecret("test-link-sign-secret-0123456789ab"),
				app.WithShareSignSecret("test-share-sign-secret-0123456789"),
				app.WithDBRegistry(root.dbReg),
				app.WithEmailRegistry(root.emailReg),
				app.WithDLQRegistry(root.dlqReg),
				app.WithTasksRegistry(root.tasksReg),
				app.WithAPISecret("test-api-secret-0123456789abcdef"),
				app.WithRouterRegistry(root.routerReg),
			)
			require.NoError(t, err)
			t.Cleanup(srv.Close)

			ts := httptest.NewServer(srv.Router)
			t.Cleanup(ts.Close)

			for _, path := range []string{"/", "/login"} {
				resp, err := ts.Client().Get(ts.URL + path)
				require.NoError(t, err)
				_ = resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode, "GET %s", path)
			}
		})
	}

	t.Run("partial_ddl_crash_and_retry", func(t *testing.T) {
		t.Run("table_exists_without_indexes", func(t *testing.T) {
			ctx := t.Context()
			h := newMySQLUpgradeHarness(t, dsn, "partial_ok", "latin1")
			h.loadV98Schema(t, true)

			// Simulate crash mid-0099 after CREATE TABLE IF NOT EXISTS pending_actions,
			// before indexes or version bump.
			_, err := h.appConn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS pending_actions (
				id VARCHAR(128) NOT NULL PRIMARY KEY,
				uid INT NOT NULL,
				browser_id VARCHAR(128) NOT NULL,
				action_type VARCHAR(128) NOT NULL,
				form_data MEDIUMTEXT NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
				expires_at TIMESTAMP NOT NULL,
				consumed_at TIMESTAMP NULL,
				FOREIGN KEY (uid) REFERENCES users(idusers) ON DELETE CASCADE
			)`)
			require.NoError(t, err)

			require.NoError(t, h.runMigrate(t, true))
			var version, maxVer int
			require.NoError(t, h.appConn.QueryRowContext(ctx, "SELECT version FROM schema_version").Scan(&version))
			require.Equal(t, handlers.ExpectedSchemaVersion, version)
			require.NoError(t, h.appConn.QueryRowContext(ctx, "SELECT MAX(version_id) FROM goose_db_version WHERE is_applied = 1").Scan(&maxVer))
			require.Equal(t, handlers.ExpectedSchemaVersion, maxVer)
		})

		t.Run("table_and_index_exist_fails_loudly", func(t *testing.T) {
			ctx := t.Context()
			h := newMySQLUpgradeHarness(t, dsn, "partial_fail", "latin1")
			h.loadV98Schema(t, true)

			_, err := h.migConn.ExecContext(ctx, `CREATE TABLE pending_actions (
				id VARCHAR(128) NOT NULL PRIMARY KEY,
				uid INT NOT NULL,
				browser_id VARCHAR(128) NOT NULL,
				action_type VARCHAR(128) NOT NULL,
				form_data MEDIUMTEXT NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
				expires_at TIMESTAMP NOT NULL,
				consumed_at TIMESTAMP NULL,
				FOREIGN KEY (uid) REFERENCES users(idusers) ON DELETE CASCADE
			);
			CREATE INDEX idx_pending_actions_expires_at ON pending_actions (expires_at);`)
			require.NoError(t, err)

			err = h.runMigrate(t, true)
			require.Error(t, err)
			require.Contains(t, err.Error(), "idx_pending_actions_expires_at")

			// Goose and schema_version must never silently advance on partial DDL failure.
			var version, maxVer int
			require.NoError(t, h.appConn.QueryRowContext(ctx, "SELECT version FROM schema_version").Scan(&version))
			require.Equal(t, 98, version)
			require.NoError(t, h.appConn.QueryRowContext(ctx, "SELECT MAX(version_id) FROM goose_db_version WHERE is_applied = 1").Scan(&maxVer))
			require.Equal(t, 98, maxVer)
		})
	})
}
