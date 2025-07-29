//go:build sqlite

package dbstart

import (
	"database/sql"
	_ "embed"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed testdata/original.sqlite.sql
var seedOriginalSchema string

//go:embed testdata/seed.sql
var seedSQL string

func execSQLSeed(t *testing.T, db *sql.DB, sqlText string) {
	execSQL(t, db, sqlText)
}

func TestSeedSQLite(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	execSQLSeed(t, db, seedOriginalSchema)
	if _, err := db.Exec("INSERT INTO schema_version (version) VALUES (1)"); err != nil {
		t.Fatalf("insert version: %v", err)
	}

	execSQLSeed(t, db, seedSQL)

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM roles").Scan(&count); err != nil {
		t.Fatalf("count roles: %v", err)
	}
	if count < 4 {
		t.Fatalf("expected roles inserted, got %d", count)
	}
}
