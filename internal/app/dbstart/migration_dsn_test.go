package dbstart

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	sqlmysql "github.com/go-sql-driver/mysql"
)

func TestMigrationConnectionString(t *testing.T) {
	const appDSN = "app:pass@tcp(127.0.0.1:3306)/goa4web?parseTime=true"
	const migrationDSN = "migrator:secret@tcp(127.0.0.1:3306)/goa4web?parseTime=true&multiStatements=true"

	cfg := &config.RuntimeConfig{DBConn: appDSN}
	if got := MigrationConnectionString(cfg); got != appDSN {
		t.Fatalf("fallback DSN = %q; want unchanged application DSN", got)
	}
	cfg.DBMigrationConn = migrationDSN
	if got := MigrationConnectionString(cfg); got != migrationDSN {
		t.Fatal("explicit migration DSN was not used verbatim")
	}
	if cfg.DBConn != appDSN || cfg.DBMigrationConn != migrationDSN {
		t.Fatal("migration connection selection modified user configuration")
	}
}

func TestValidateMigrationConnection(t *testing.T) {
	const dsn = "app:pass@tcp(127.0.0.1:3306)/goa4web?parseTime=true"
	for _, tc := range []struct {
		name     string
		version  int
		wantFail bool
	}{
		{name: "pending migrations", version: handlers.ExpectedSchemaVersion - 1, wantFail: true},
		{name: "already migrated", version: handlers.ExpectedSchemaVersion},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = db.Close() }()
			mock.ExpectQuery("SELECT version FROM schema_version").WillReturnRows(
				sqlmock.NewRows([]string{"version"}).AddRow(tc.version))
			err = ValidateMigrationConnection(context.Background(), db, "mysql", dsn)
			if tc.wantFail {
				if err == nil || !strings.Contains(err.Error(), "DB_MIGRATION_CONN") {
					t.Fatalf("expected actionable configuration error, got %v", err)
				}
			} else if err != nil {
				t.Fatalf("already-migrated database should start with original DSN: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if err := ValidateMigrationConnection(context.Background(), db, "mysql", dsn+"&multiStatements=true"); err != nil {
		t.Fatalf("explicit multiStatements=true should be accepted: %v", err)
	}
	if err := ValidateMigrationConnection(context.Background(), db, "sqlite", "file:db?mode=rwc"); err != nil {
		t.Fatalf("SQLite should not require MySQL multi-statements: %v", err)
	}
	if err := ValidateMigrationConnection(context.Background(), db, "mysql", "not a valid DSN"); err == nil {
		t.Fatal("invalid MySQL DSN should fail validation")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMigrationConnectionAgainstMariaDB(t *testing.T) {
	dsn := os.Getenv("GOA4WEB_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set GOA4WEB_TEST_MYSQL_DSN to exercise a migration connection against MariaDB")
	}
	cfg, err := sqlmysql.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.MultiStatements = false
	applicationDSN := cfg.FormatDSN()
	cfg.MultiStatements = true
	explicitDSN := cfg.FormatDSN()

	runtimeCfg := &config.RuntimeConfig{DBConn: applicationDSN, DBMigrationConn: explicitDSN}
	chosenDSN := MigrationConnectionString(runtimeCfg)
	if chosenDSN != explicitDSN || runtimeCfg.DBConn != applicationDSN {
		t.Fatal("migration DSN selection changed the application's original DSN")
	}
	db, err := sql.Open("mysql", chosenDSN)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	rows, err := db.Query("SELECT 1; SELECT 2")
	if err != nil {
		t.Fatalf("explicit migration DSN rejected multi-statement SQL: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var first, second int
	if !rows.Next() {
		t.Fatal("missing first statement result")
	}
	if err := rows.Scan(&first); err != nil {
		t.Fatal(err)
	}
	if !rows.NextResultSet() || !rows.Next() {
		t.Fatal("missing second statement result")
	}
	if err := rows.Scan(&second); err != nil {
		t.Fatal(err)
	}
	if first != 1 || second != 2 {
		t.Fatalf("statement results = (%d, %d), want (1, 2)", first, second)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}
