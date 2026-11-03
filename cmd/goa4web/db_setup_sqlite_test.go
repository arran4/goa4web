//go:build sqlite

package main

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/app/dbstart"
	_ "modernc.org/sqlite"
)

func TestDBSetupInitializesSQLiteDatabase(t *testing.T) {
	dsn := fmt.Sprintf("file:db_setup_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	anchor, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open SQLite anchor: %v", err)
	}
	defer func() { _ = anchor.Close() }()
	if err := anchor.Ping(); err != nil {
		t.Fatalf("connect SQLite anchor: %v", err)
	}

	root, err := parseRoot([]string{"goa4web", "--db-driver", "sqlite3", "--db-conn", dsn, "db", "setup"})
	if err != nil {
		t.Fatalf("parse root command: %v", err)
	}
	if err := root.Run(); err != nil {
		t.Fatalf("run db setup: %v", err)
	}
	defer root.Close()
	if _, err := root.DB(); err != nil {
		t.Fatalf("open initialized SQLite database normally: %v", err)
	}

	if err := dbstart.EnsureSchemaWithDriver(context.Background(), anchor, "sqlite3"); err != nil {
		t.Fatalf("verify schema version: %v", err)
	}
	var roleCount, languageID int
	if err := anchor.QueryRow("SELECT COUNT(*) FROM roles").Scan(&roleCount); err != nil {
		t.Fatalf("count seeded roles: %v", err)
	}
	if roleCount == 0 {
		t.Fatal("expected seeded roles")
	}
	if err := anchor.QueryRow("SELECT id FROM language WHERE nameof = 'English'").Scan(&languageID); err != nil {
		t.Fatalf("read default language: %v", err)
	}
	if languageID != 1 {
		t.Fatalf("default language ID = %d, want 1", languageID)
	}
}
