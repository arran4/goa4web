//go:build sqlite || sqlite3

package db

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSystemCreateNotificationForActivitySQLiteKeepsHighestPriority(t *testing.T) {
	ctx := context.Background()
	conn, err := sql.Open("sqlite", "file:notification_activity_upsert?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.ExecContext(ctx, `
		CREATE TABLE notifications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			users_idusers INTEGER NOT NULL,
			link TEXT,
			message TEXT,
			activity_key TEXT,
			priority INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			read_at DATETIME,
			UNIQUE (users_idusers, activity_key)
		)`); err != nil {
		t.Fatalf("create notifications table: %v", err)
	}

	queries := NewForDriver(conn, "sqlite3")
	activity := sql.NullString{String: "reply:comment:42", Valid: true}
	if err := queries.SystemCreateNotificationForActivity(ctx, SystemCreateNotificationForActivityParams{
		RecipientID: 7,
		Link:        sql.NullString{String: "/forum/topic/1/thread/2#c42", Valid: true},
		Message:     sql.NullString{String: "subscriber contribution", Valid: true},
		ActivityKey: activity,
		Priority:    20,
	}); err != nil {
		t.Fatalf("insert subscriber contribution: %v", err)
	}
	if err := queries.SystemCreateNotificationForActivity(ctx, SystemCreateNotificationForActivityParams{
		RecipientID: 7,
		Message:     sql.NullString{String: "admin contribution", Valid: true},
		ActivityKey: activity,
		Priority:    10,
	}); err != nil {
		t.Fatalf("upsert lower-priority contribution: %v", err)
	}

	var count, priority int
	var message string
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*), MAX(priority), MAX(message) FROM notifications WHERE users_idusers = 7`).Scan(&count, &priority, &message); err != nil {
		t.Fatalf("read retained notification: %v", err)
	}
	if count != 1 || priority != 20 || message != "subscriber contribution" {
		t.Fatalf("retained notification = count %d, priority %d, message %q", count, priority, message)
	}
}
