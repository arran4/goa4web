package dbsqlite

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestAcknowledgeDigestNotificationsOnlyUpdatesStoredSnapshot(t *testing.T) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	defer func() { _ = database.Close() }()

	for _, statement := range []string{
		"CREATE TABLE notifications (id INTEGER PRIMARY KEY, users_idusers INTEGER NOT NULL, read_at DATETIME)",
		"CREATE TABLE digest_deliveries (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL, dismiss_notifications INTEGER NOT NULL, acknowledged_at DATETIME)",
		"CREATE TABLE digest_delivery_notifications (digest_delivery_id INTEGER NOT NULL, notification_id INTEGER NOT NULL)",
		"CREATE TABLE pending_emails (id INTEGER PRIMARY KEY, digest_delivery_id INTEGER)",
		"INSERT INTO notifications (id, users_idusers) VALUES (1, 7), (2, 7)",
		"INSERT INTO digest_deliveries (id, user_id, dismiss_notifications) VALUES (10, 7, 1)",
		"INSERT INTO digest_delivery_notifications (digest_delivery_id, notification_id) VALUES (10, 1)",
		"INSERT INTO pending_emails (id, digest_delivery_id) VALUES (20, 10)",
	} {
		if _, err := database.Exec(statement); err != nil {
			t.Fatalf("setup %q: %v", statement, err)
		}
	}

	if err := New(database).SystemAcknowledgeDigestNotificationsForPendingEmail(context.Background(), 20); err != nil {
		t.Fatalf("acknowledge digest: %v", err)
	}
	var snapshotRead, laterRead sql.NullTime
	if err := database.QueryRow("SELECT read_at FROM notifications WHERE id = 1").Scan(&snapshotRead); err != nil {
		t.Fatalf("read snapshot notification: %v", err)
	}
	if err := database.QueryRow("SELECT read_at FROM notifications WHERE id = 2").Scan(&laterRead); err != nil {
		t.Fatalf("read later notification: %v", err)
	}
	if !snapshotRead.Valid {
		t.Fatal("snapshot notification was not acknowledged")
	}
	if laterRead.Valid {
		t.Fatal("notification created after the digest snapshot was dismissed")
	}
}
