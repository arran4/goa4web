package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/arran4/goa4web/handlers"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func TestCanonicalMySQLInstallCanRunMigrations(t *testing.T) {
	dsn := os.Getenv("GOA4WEB_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set GOA4WEB_TEST_MYSQL_DSN to test the MySQL install and migration commands")
	}
	for _, command := range []string{"create", "setup"} {
		t.Run(command, func(t *testing.T) {
			ctx := t.Context()
			cfg, err := mysql.ParseDSN(dsn)
			require.NoError(t, err)
			cfg.DBName = ""
			admin, err := sql.Open("mysql", cfg.FormatDSN())
			require.NoError(t, err)
			t.Cleanup(func() { _ = admin.Close() })

			name := fmt.Sprintf("goa4web_install_%d", time.Now().UnixNano())
			_, err = admin.ExecContext(ctx, "CREATE DATABASE `"+name+"` DEFAULT CHARACTER SET utf8mb4")
			require.NoError(t, err)
			t.Cleanup(func() { _, _ = admin.Exec("DROP DATABASE `" + name + "`") })
			cfg.DBName = name
			cfg.MultiStatements = false
			appDSN := cfg.FormatDSN()
			cfg.MultiStatements = true
			migrationDSN := cfg.FormatDSN()

			run := func(command string) error {
				root, err := parseRoot([]string{"goa4web", "--db-driver", "mysql", "--db-conn", appDSN,
					"--db-migration-conn", migrationDSN, "db", command})
				if err != nil {
					return err
				}
				defer root.Close()
				return root.Run()
			}
			require.NoError(t, run(command))

			conn, err := sql.Open("mysql", appDSN)
			require.NoError(t, err)
			t.Cleanup(func() { _ = conn.Close() })
			_, err = conn.ExecContext(ctx, "INSERT INTO users (username) VALUES ('snapshot_fixture')")
			require.NoError(t, err)
			require.NoError(t, run("migrate"))
			require.NoError(t, run("migrate"))
			var username string
			require.NoError(t, conn.QueryRowContext(ctx, "SELECT username FROM users WHERE username='snapshot_fixture'").Scan(&username))
			require.Equal(t, "snapshot_fixture", username)
			var count, version int
			err = conn.QueryRowContext(ctx, "SELECT COUNT(DISTINCT version_id), MAX(version_id) FROM goose_db_version WHERE is_applied=1").Scan(&count, &version)
			require.NoError(t, err)
			require.Equal(t, handlers.ExpectedSchemaVersion, version)
			require.Equal(t, version+1, count)
		})
	}
}
