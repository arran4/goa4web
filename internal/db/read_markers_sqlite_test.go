//go:build sqlite || sqlite3

package db

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestUpsertContentReadMarkerSQLiteDoesNotRegressPosition(t *testing.T) {
	ctx := context.Background()
	conn, err := sql.Open("sqlite", "file:read_marker_upsert?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.ExecContext(ctx, `
		CREATE TABLE content_read_markers (
			item TEXT NOT NULL,
			item_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			last_comment_id INTEGER NOT NULL,
			UNIQUE (user_id, item_id, item)
		)`); err != nil {
		t.Fatalf("create content_read_markers table: %v", err)
	}

	queries := NewForDriver(conn, "sqlite3")
	newer := UpsertContentReadMarkerParams{Item: "thread", ItemID: 2, UserID: 7, LastCommentID: 42}
	if err := queries.UpsertContentReadMarker(ctx, newer); err != nil {
		t.Fatalf("save newer marker: %v", err)
	}
	older := newer
	older.LastCommentID = 21
	if err := queries.UpsertContentReadMarker(ctx, older); err != nil {
		t.Fatalf("save delayed older marker: %v", err)
	}

	marker, err := queries.GetContentReadMarker(ctx, GetContentReadMarkerParams{Item: "thread", ItemID: 2, UserID: 7})
	if err != nil {
		t.Fatalf("load marker: %v", err)
	}
	if marker.LastCommentID != newer.LastCommentID {
		t.Fatalf("last comment ID = %d, want retained newer position %d", marker.LastCommentID, newer.LastCommentID)
	}
}
