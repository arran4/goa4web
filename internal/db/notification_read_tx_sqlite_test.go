//go:build sqlite || sqlite3

package db

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestNotificationReadStateLinearizabilitySQLite(t *testing.T) {
	ctx := context.Background()
	conn, err := sql.Open("sqlite", "file:"+t.Name()+"?mode=memory&cache=shared&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("open SQLite database: %v", err)
	}
	conn.SetMaxOpenConns(4)
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.ExecContext(ctx, `
		CREATE TABLE content_read_markers (
			id integer PRIMARY KEY AUTOINCREMENT,
			item TEXT NOT NULL,
			item_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			last_comment_id INTEGER NOT NULL,
			last_activity_position INTEGER NOT NULL DEFAULT 0,
			UNIQUE (user_id, item_id, item)
		);
		CREATE TABLE notifications (
			id integer PRIMARY KEY AUTOINCREMENT,
			users_idusers integer NOT NULL,
			link TEXT,
			message TEXT,
			activity_key TEXT,
			priority integer NOT NULL DEFAULT 0,
			resource_type TEXT,
			resource_id integer,
			activity_position integer,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			read_at TIMESTAMP DEFAULT NULL,
			UNIQUE (users_idusers, activity_key)
		);
	`); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	queries := NewForDriver(conn, "sqlite3")
	runNotificationReadStateLinearizabilitySuite(t, queries)
}
