//go:build sqlite

package dbstart

import (
	"context"
	"database/sql"
	_ "embed"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/migrations"
)

//go:embed testdata/original.sqlite.sql
var migOriginalSchema string

func execSQLMig(t *testing.T, db *sql.DB, sqlText string) {
	execSQL(t, db, sqlText)
}

func TestMigrationsSQLite(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	execSQLMig(t, db, migOriginalSchema)
	if _, err := db.Exec("INSERT INTO schema_version (version) VALUES (1)"); err != nil {
		t.Fatalf("insert version: %v", err)
	}

	if err := Apply(context.Background(), db, migrations.FS, false, "sqlite"); err != nil {
		t.Fatalf("apply: %v", err)
	}

	var version int
	if err := db.QueryRow("SELECT version FROM schema_version").Scan(&version); err != nil {
		t.Fatalf("select version: %v", err)
	}
	if version != handlers.ExpectedSchemaVersion {
		t.Fatalf("expected version %d got %d", handlers.ExpectedSchemaVersion, version)
	}
}
