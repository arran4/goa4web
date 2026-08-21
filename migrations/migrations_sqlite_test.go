//go:build sqlite || sqlite3

package migrations_test

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/migrations"
	_ "modernc.org/sqlite"
)

func TestSQLiteMigrationsStepByStep(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_step.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open test sqlite db: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	for v := 1; v <= handlers.ExpectedSchemaVersion; v++ {
		filename := fmt.Sprintf("%04d_sqlite.sql", v)
		content, err := migrations.FS.ReadFile(filename)
		if err != nil {
			t.Fatalf("read migration %s: %v", filename, err)
		}

		rawSQL := string(content)
		// Only take the Up section
		if idx := strings.Index(rawSQL, "-- +goose Down"); idx != -1 {
			rawSQL = rawSQL[:idx]
		}
		lines := strings.Split(rawSQL, "\n")
		var cleanLines []string
		for _, l := range lines {
			trimmed := strings.TrimSpace(l)
			if strings.HasPrefix(trimmed, "-- +goose") {
				continue
			}
			cleanLines = append(cleanLines, l)
		}
		cleanSQL := strings.Join(cleanLines, "\n")

		// Execute migration
		if _, err := db.ExecContext(ctx, cleanSQL); err != nil {
			t.Fatalf("Migration %s failed: %v\nSQL was:\n%s", filename, err, cleanSQL)
		}
	}
}

func TestSQLiteMigrationsExecuteFromScratch(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_migration.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open test sqlite db: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	if err := dbstart.Apply(ctx, db, migrations.FS, testing.Verbose(), "sqlite3"); err != nil {
		t.Fatalf("failed to apply sqlite migrations: %v", err)
	}

	version, err := dbstart.SchemaVersionWithDriver(ctx, db, "sqlite3")
	if err != nil {
		t.Fatalf("failed to read sqlite schema version: %v", err)
	}

	if version != handlers.ExpectedSchemaVersion {
		t.Fatalf("expected schema version %d, got %d", handlers.ExpectedSchemaVersion, version)
	}

	// Verify key tables were created
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		t.Fatalf("failed to query users table: %v", err)
	}

	// Verify migration 0095 columns in forumthread
	var replyToCommentID, replyToThreadID sql.NullInt64
	err = db.QueryRowContext(ctx, "SELECT reply_to_comment_id, reply_to_thread_id FROM forumthread LIMIT 1").Scan(&replyToCommentID, &replyToThreadID)
	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("failed to query 0095 forumthread columns: %v", err)
	}

	// Verify long URLs > 255 chars with identical first 255 chars in SQLite
	prefix := "https://example.com/very/long/path/" + strings.Repeat("a", 230) + "?common=true&suffix="
	url1 := prefix + "one"
	url2 := prefix + "two"

	if _, err := db.ExecContext(ctx, `INSERT INTO external_links (url, clicks) VALUES (?, 10)`, url1); err != nil {
		t.Fatalf("insert long sqlite url 1: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO external_links (url, clicks) VALUES (?, 20)`, url2); err != nil {
		t.Fatalf("insert long sqlite url 2: %v", err)
	}

	var id1, id2, clicks1, clicks2 int
	if err := db.QueryRowContext(ctx, `SELECT id, clicks FROM external_links WHERE url = ?`, url1).Scan(&id1, &clicks1); err != nil {
		t.Fatalf("query long sqlite url 1: %v", err)
	}
	if err := db.QueryRowContext(ctx, `SELECT id, clicks FROM external_links WHERE url = ?`, url2).Scan(&id2, &clicks2); err != nil {
		t.Fatalf("query long sqlite url 2: %v", err)
	}
	if id1 == id2 || clicks1 != 10 || clicks2 != 20 {
		t.Fatalf("sqlite long urls collided: id1=%d, id2=%d, clicks1=%d, clicks2=%d", id1, id2, clicks1, clicks2)
	}
}
