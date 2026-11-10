package dbsqlite_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/internal/dbsqlite"
	"github.com/arran4/goa4web/migrations"
	_ "modernc.org/sqlite"
)

func TestImageCacheEntryUpsertsUseSQLiteConflictSyntax(t *testing.T) {
	ctx := context.Background()
	conn, err := sql.Open("sqlite", "file:image_cache_upsert_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if err := dbstart.Apply(ctx, conn, migrations.FS, false, "sqlite3"); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	queries := dbsqlite.New(conn)
	now := time.Date(2026, time.August, 1, 9, 0, 0, 0, time.UTC)
	source := sql.NullString{String: "/ab/cd/example.jpg", Valid: true}
	if err := queries.UpsertImageCacheEntry(ctx, dbsqlite.UpsertImageCacheEntryParams{
		ID:         "example_thumb_256x128.jpg",
		SourceUrl:  source,
		SourceKind: "uploaded",
		Status:     "ready",
		CreatedAt:  now,
		LastUsedAt: sql.NullTime{Time: now, Valid: true},
		Width:      sql.NullInt64{Int64: 128, Valid: true},
		Height:     sql.NullInt64{Int64: 96, Valid: true},
	}); err != nil {
		t.Fatalf("insert image cache entry: %v", err)
	}
	if err := queries.UpsertImageCacheEntry(ctx, dbsqlite.UpsertImageCacheEntryParams{
		ID:         "example_thumb_256x128.jpg",
		SourceUrl:  source,
		SourceKind: "uploaded",
		Status:     "ready",
		CreatedAt:  now,
		LastUsedAt: sql.NullTime{Time: now.Add(time.Minute), Valid: true},
		Width:      sql.NullInt64{Int64: 256, Valid: true},
		Height:     sql.NullInt64{Int64: 192, Valid: true},
	}); err != nil {
		t.Fatalf("update image cache entry: %v", err)
	}
	entry, err := queries.GetImageCacheEntry(ctx, "example_thumb_256x128.jpg")
	if err != nil {
		t.Fatalf("get image cache entry: %v", err)
	}
	if !entry.Width.Valid || entry.Width.Int64 != 256 || !entry.Height.Valid || entry.Height.Int64 != 192 {
		t.Errorf("image cache dimensions = %v x %v, want 256 x 192", entry.Width, entry.Height)
	}

	if err := queries.CreatePendingImageCacheEntry(ctx, dbsqlite.CreatePendingImageCacheEntryParams{
		ID:            "pending-image",
		SourceUrl:     sql.NullString{String: "https://example.test/image.jpg", Valid: true},
		SourceKind:    "remote",
		CreatedAt:     now,
		LastUsedAt:    sql.NullTime{Time: now, Valid: true},
		NextAttemptAt: sql.NullTime{Time: now.Add(time.Hour), Valid: true},
	}); err != nil {
		t.Fatalf("insert pending image cache entry: %v", err)
	}

	threadImage := sql.NullString{String: "/ab/cd/example.jpg", Valid: true}
	if err := queries.CreateThreadImage(ctx, dbsqlite.CreateThreadImageParams{
		ThreadID: 42,
		Path:     threadImage,
	}); err != nil {
		t.Fatalf("insert thread image: %v", err)
	}
	if err := queries.CreateThreadImage(ctx, dbsqlite.CreateThreadImageParams{
		ThreadID: 42,
		Path:     threadImage,
	}); err != nil {
		t.Fatalf("repeat thread image insert: %v", err)
	}
	paths, err := queries.ListThreadImagePaths(ctx, dbsqlite.ListThreadImagePathsParams{
		ForumthreadID: 42,
		Paths:         []sql.NullString{threadImage},
	})
	if err != nil {
		t.Fatalf("list thread image: %v", err)
	}
	if len(paths) != 1 || !paths[0].Valid || paths[0].String != threadImage.String {
		t.Errorf("thread image paths = %#v, want %#v", paths, []sql.NullString{threadImage})
	}
}
