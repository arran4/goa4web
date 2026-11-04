package dbstart

import (
	"database/sql"
	"os"
	"testing"

	sqlmysql "github.com/go-sql-driver/mysql"
)

func TestMigrationConnectionString(t *testing.T) {
	const normalDSN = "test:secret@tcp(127.0.0.1:3306)/goa4web?parseTime=true&multiStatements=false"

	normal, err := sqlmysql.ParseDSN(normalDSN)
	if err != nil {
		t.Fatal(err)
	}
	if normal.MultiStatements {
		t.Fatal("test setup: application DSN unexpectedly permits multiple statements")
	}

	migrationDSN, err := MigrationConnectionString("mysql", normalDSN)
	if err != nil {
		t.Fatal(err)
	}
	migration, err := sqlmysql.ParseDSN(migrationDSN)
	if err != nil {
		t.Fatal(err)
	}
	if !migration.MultiStatements {
		t.Fatal("migration DSN must permit multi-statement Goose blocks")
	}
	if migration.User != normal.User || migration.Passwd != normal.Passwd ||
		migration.DBName != normal.DBName || migration.Net != normal.Net ||
		migration.Addr != normal.Addr || migration.ParseTime != normal.ParseTime {
		t.Fatal("migration DSN changed settings other than multiStatements")
	}
	if normal.MultiStatements {
		t.Fatal("the application DSN was mutated")
	}

	sqliteDSN := "file:goa4web.db?mode=rwc"
	got, err := MigrationConnectionString("sqlite", sqliteDSN)
	if err != nil || got != sqliteDSN {
		t.Fatalf("SQLite DSN = %q, err = %v; want unchanged", got, err)
	}

	if _, err := MigrationConnectionString("mysql", "not a valid DSN"); err == nil {
		t.Fatal("invalid MySQL migration DSN should fail")
	}
}

func TestMigrationConnectionAgainstMariaDB(t *testing.T) {
	dsn := os.Getenv("GOA4WEB_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set GOA4WEB_TEST_MYSQL_DSN to exercise migration connection against MariaDB")
	}

	cfg, err := sqlmysql.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	// Start from the production behaviour, not the multiStatements=true DSN
	// used by other SQL-integration tests in CI.
	cfg.MultiStatements = false
	migrationDSN, err := MigrationConnectionString("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("mysql", migrationDSN)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT 1; SELECT 2")
	if err != nil {
		t.Fatalf("migration connection rejected multi-statement SQL: %v", err)
	}
	defer rows.Close()
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
