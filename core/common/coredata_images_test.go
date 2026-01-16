package common

import (
	"context"
	"database/sql"
	"path"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestCreateCommentValidatesGalleryImages(t *testing.T) {
	imageID := "abcd1234.jpg"
	imagePath := path.Join("/uploads", imageID[:2], imageID[2:4], imageID)
	text := "[img image:" + imageID + "]"

	t.Run("accepts gallery image and records thread usage", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.ListUploadedImagePathsByUserFn = func(ctx context.Context, arg db.ListUploadedImagePathsByUserParams) ([]sql.NullString, error) {
			return []sql.NullString{{String: imagePath, Valid: true}}, nil
		}
		queries.CreateCommentInSectionForCommenterResult = 42
		cd := NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
		commentID, err := cd.CreateCommentInSectionForCommenter("forum", "topic", 1, 1, 9, 1, text)
		if err != nil {
			t.Fatalf("CreateCommentInSectionForCommenter: %v", err)
		}
		if commentID != 42 {
			t.Fatalf("comment id = %d, want 42", commentID)
		}
		if len(queries.ListUploadedImagePathsByUserCalls) != 1 {
			t.Fatalf("expected gallery lookup, got %d calls", len(queries.ListUploadedImagePathsByUserCalls))
		}
		if len(queries.CreateCommentInSectionForCommenterCalls) != 1 {
			t.Fatalf("expected comment creation, got %d calls", len(queries.CreateCommentInSectionForCommenterCalls))
		}
		gotPaths := queries.ListUploadedImagePathsByUserCalls[0].Paths
		if len(gotPaths) != 1 || !gotPaths[0].Valid || gotPaths[0].String != imagePath {
			t.Fatalf("paths = %#v, want %q", gotPaths, imagePath)
		}
		if len(queries.CreateThreadImageCalls) != 1 {
			t.Fatalf("expected thread image record, got %d calls", len(queries.CreateThreadImageCalls))
		}
	})

	t.Run("accepts thread image", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.ListUploadedImagePathsByUserFn = func(ctx context.Context, arg db.ListUploadedImagePathsByUserParams) ([]sql.NullString, error) {
			return []sql.NullString{}, nil
		}
		queries.ListThreadImagePathsFn = func(ctx context.Context, arg db.ListThreadImagePathsParams) ([]sql.NullString, error) {
			return []sql.NullString{{String: imagePath, Valid: true}}, nil
		}
		queries.CreateCommentInSectionForCommenterResult = 42
		cd := NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
		if _, err := cd.CreateCommentInSectionForCommenter("forum", "topic", 1, 1, 9, 1, text); err != nil {
			t.Fatalf("expected thread image acceptance: %v", err)
		}
		if len(queries.ListThreadImagePathsCalls) != 1 {
			t.Fatalf("expected thread image lookup, got %d calls", len(queries.ListThreadImagePathsCalls))
		}
		if len(queries.CreateThreadImageCalls) != 1 {
			t.Fatalf("expected thread image record, got %d calls", len(queries.CreateThreadImageCalls))
		}
	})

	t.Run("rejects missing gallery image", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.ListUploadedImagePathsByUserFn = func(ctx context.Context, arg db.ListUploadedImagePathsByUserParams) ([]sql.NullString, error) {
			return []sql.NullString{}, nil
		}
		queries.ListThreadImagePathsFn = func(ctx context.Context, arg db.ListThreadImagePathsParams) ([]sql.NullString, error) {
			return []sql.NullString{}, nil
		}
		cd := NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
		if _, err := cd.CreateCommentInSectionForCommenter("forum", "topic", 1, 1, 9, 1, text); err == nil {
			t.Fatal("expected error for missing gallery image")
		}
		if len(queries.CreateCommentInSectionForCommenterCalls) != 0 {
			t.Fatalf("expected no comment creation, got %d calls", len(queries.CreateCommentInSectionForCommenterCalls))
		}
	})
}
