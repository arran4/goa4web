package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/migrations"
	"github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

func TestCanonicalMySQLInstallCanRunMigrations(t *testing.T) {
	dsn := os.Getenv("GOA4WEB_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set GOA4WEB_TEST_MYSQL_DSN to test the MySQL install and migration commands")
	}
	for _, command := range []string{"create", "setup"} {
		for _, charset := range []string{"latin1", "utf8mb4"} {
			t.Run(command+"/"+charset, func(t *testing.T) {
				ctx := t.Context()
				cfg, err := mysql.ParseDSN(dsn)
				require.NoError(t, err)
				cfg.DBName = ""
				admin, err := sql.Open("mysql", cfg.FormatDSN())
				require.NoError(t, err)
				t.Cleanup(func() { _ = admin.Close() })

				name := fmt.Sprintf("goa4web_install_%d", time.Now().UnixNano())
				_, err = admin.ExecContext(ctx, "CREATE DATABASE `"+name+"` DEFAULT CHARACTER SET "+charset)
				require.NoError(t, err)
				t.Cleanup(func() { _, _ = admin.Exec("DROP DATABASE `" + name + "`") })
				cfg.DBName = name
				cfg.MultiStatements = false
				appDSN := cfg.FormatDSN()
				cfg.MultiStatements = true
				migrationDSN := cfg.FormatDSN()

				run := func(cmd string, withMigrationConn bool) error {
					args := []string{"goa4web", "--db-driver", "mysql", "--db-conn", appDSN}
					if withMigrationConn {
						args = append(args, "--db-migration-conn", migrationDSN)
					}
					args = append(args, "db", cmd)
					root, err := parseRoot(args)
					if err != nil {
						return err
					}
					defer root.Close()
					return root.Run()
				}
				require.NoError(t, run(command, true))

				conn, err := sql.Open("mysql", appDSN)
				require.NoError(t, err)
				t.Cleanup(func() { _ = conn.Close() })

				rows, err := conn.QueryContext(ctx, "SELECT 1; SELECT 2")
				if err == nil {
					_ = rows.Close()
				}
				require.Error(t, err, "application DSN must disallow multi-statement SQL")

				_, err = conn.ExecContext(ctx, "INSERT INTO users (username) VALUES ('snapshot_fixture')")
				require.NoError(t, err)

				require.NoError(t, run("migrate", true))
				// Already-current databases must not require multiStatements=true on migrate.
				require.NoError(t, run("migrate", false))

				var username string
				require.NoError(t, conn.QueryRowContext(ctx, "SELECT username FROM users WHERE username='snapshot_fixture'").Scan(&username))
				require.Equal(t, "snapshot_fixture", username)

				var legacyVersion int
				require.NoError(t, conn.QueryRowContext(ctx, "SELECT version FROM schema_version").Scan(&legacyVersion))
				require.Equal(t, handlers.ExpectedSchemaVersion, legacyVersion)

				var count, minVer, version int
				err = conn.QueryRowContext(ctx, "SELECT COUNT(DISTINCT version_id), MIN(version_id), MAX(version_id) FROM goose_db_version WHERE is_applied=1").Scan(&count, &minVer, &version)
				require.NoError(t, err)
				require.Equal(t, 0, minVer)
				require.Equal(t, handlers.ExpectedSchemaVersion, version)
				require.Equal(t, version+1, count)

				// Prove a synthetic follow-on migration executes cleanly after fresh install.
				migFS := makeSyntheticFollowOnFS(t, handlers.ExpectedSchemaVersion+1)
				migConn, err := sql.Open("mysql", migrationDSN)
				require.NoError(t, err)
				t.Cleanup(func() { _ = migConn.Close() })
				provider, err := goose.NewProvider(goose.DialectMySQL, migConn, migFS)
				require.NoError(t, err)
				_, err = provider.Up(ctx)
				require.NoError(t, err)
				gotVer, err := provider.GetDBVersion(ctx)
				require.NoError(t, err)
				require.Equal(t, int64(handlers.ExpectedSchemaVersion+1), gotVer)
				require.NoError(t, conn.QueryRowContext(ctx, "SELECT version FROM schema_version").Scan(&legacyVersion))
				require.Equal(t, handlers.ExpectedSchemaVersion+1, legacyVersion)
			})
		}
	}
}

func makeSyntheticFollowOnFS(t *testing.T, nextVersion int) fs.FS {
	t.Helper()
	base := migrations.ForDriver("mysql")
	entries, err := fs.ReadDir(base, ".")
	require.NoError(t, err)
	overlay := fstest.MapFS{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		data, err := fs.ReadFile(base, entry.Name())
		require.NoError(t, err)
		overlay[entry.Name()] = &fstest.MapFile{Data: data}
	}
	name := fmt.Sprintf("%04d_mysql.sql", nextVersion)
	sqlBody := fmt.Sprintf("-- +goose Up\nUPDATE schema_version SET version = %d;\n-- +goose Down\nUPDATE schema_version SET version = %d;\n", nextVersion, nextVersion-1)
	overlay[name] = &fstest.MapFile{Data: []byte(sqlBody)}
	return overlay
}
