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
			resource_type TEXT,
			resource_id INTEGER,
			activity_position INTEGER,
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

func TestSystemAcknowledgeNotificationsForResourceSQLitePreservesNewerActivity(t *testing.T) {
	ctx := context.Background()
	conn, err := sql.Open("sqlite", "file:notification_resource_acknowledgement?mode=memory&cache=shared")
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
			resource_type TEXT,
			resource_id INTEGER,
			activity_position INTEGER,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			read_at DATETIME,
			UNIQUE (users_idusers, activity_key)
		)`); err != nil {
		t.Fatalf("create notifications table: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO notifications (users_idusers, activity_key, priority, resource_type, resource_id, activity_position)
		VALUES (7, 'reply:comment:40', 20, 'thread', 2, 40),
		       (7, 'reply:comment:43', 20, 'thread', 2, 43),
		       (7, 'reply:comment:other-thread', 20, 'thread', 3, 40),
		       (8, 'reply:comment:other-user', 20, 'thread', 2, 40)`); err != nil {
		t.Fatalf("seed notifications: %v", err)
	}

	queries := NewForDriver(conn, "sqlite3")
	if err := queries.SystemAcknowledgeNotificationsForResource(ctx, SystemAcknowledgeNotificationsForResourceParams{
		RecipientID:      7,
		ResourceType:     sql.NullString{String: "thread", Valid: true},
		ResourceID:       sql.NullInt32{Int32: 2, Valid: true},
		ActivityPosition: sql.NullInt32{Int32: 42, Valid: true},
	}); err != nil {
		t.Fatalf("acknowledge notifications: %v", err)
	}

	rows, err := conn.QueryContext(ctx, `SELECT activity_key, read_at IS NOT NULL FROM notifications ORDER BY activity_key`)
	if err != nil {
		t.Fatalf("read notification state: %v", err)
	}
	defer rows.Close()
	got := map[string]bool{}
	for rows.Next() {
		var activityKey string
		var read bool
		if err := rows.Scan(&activityKey, &read); err != nil {
			t.Fatalf("scan notification state: %v", err)
		}
		got[activityKey] = read
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate notification state: %v", err)
	}
	if !got["reply:comment:40"] || got["reply:comment:43"] || got["reply:comment:other-thread"] || got["reply:comment:other-user"] {
		t.Fatalf("notification read state = %#v", got)
	}
}
