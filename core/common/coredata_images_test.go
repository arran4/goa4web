package common

import (
	"context"
	"database/sql"
	"path"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
)

func TestCreateCommentValidatesGalleryImages(t *testing.T) {
	imageID := "abcd1234.jpg"
	imagePath := path.Join("/uploads", imageID[:2], imageID[2:4], imageID)
	text := "[img image:" + imageID + "]"

	t.Run("accepts gallery image and shares with participants", func(t *testing.T) {
		queries := &db.QuerierStub{
			ListUploadedImagePathsByUserFn: func(ctx context.Context, arg db.ListUploadedImagePathsByUserParams) ([]sql.NullString, error) {
				return []sql.NullString{{String: imagePath, Valid: true}}, nil
			},
			ListThreadParticipantIDsFn: func(ctx context.Context, threadID int32) ([]int32, error) {
				return []int32{17}, nil
			},
			CreateCommentInSectionForCommenterResult: 42,
		}
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
		if len(queries.ShareUploadedImageWithUserCalls) != 2 {
			t.Fatalf("expected share for participants, got %d calls", len(queries.ShareUploadedImageWithUserCalls))
		}
	})

	t.Run("accepts thread gallery image", func(t *testing.T) {
		queries := &db.QuerierStub{
			ListUploadedImagePathsByUserFn: func(ctx context.Context, arg db.ListUploadedImagePathsByUserParams) ([]sql.NullString, error) {
				return []sql.NullString{}, nil
			},
			ListUploadedImagePathsByThreadFn: func(ctx context.Context, arg db.ListUploadedImagePathsByThreadParams) ([]sql.NullString, error) {
				return []sql.NullString{{String: imagePath, Valid: true}}, nil
			},
			ListThreadParticipantIDsFn: func(ctx context.Context, threadID int32) ([]int32, error) {
				return []int32{17}, nil
			},
			CreateCommentInSectionForCommenterResult: 42,
		}
		cd := NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
		if _, err := cd.CreateCommentInSectionForCommenter("forum", "topic", 1, 1, 9, 1, text); err != nil {
			t.Fatalf("expected thread image acceptance: %v", err)
		}
		if len(queries.ListUploadedImagePathsByThreadCalls) != 1 {
			t.Fatalf("expected thread image lookup, got %d calls", len(queries.ListUploadedImagePathsByThreadCalls))
		}
		if len(queries.ShareUploadedImageWithUserCalls) != 2 {
			t.Fatalf("expected share for participants, got %d calls", len(queries.ShareUploadedImageWithUserCalls))
		}
	})

	t.Run("rejects missing gallery image", func(t *testing.T) {
		queries := &db.QuerierStub{
			ListUploadedImagePathsByUserFn: func(ctx context.Context, arg db.ListUploadedImagePathsByUserParams) ([]sql.NullString, error) {
				return []sql.NullString{}, nil
			},
			ListUploadedImagePathsByThreadFn: func(ctx context.Context, arg db.ListUploadedImagePathsByThreadParams) ([]sql.NullString, error) {
				return []sql.NullString{}, nil
			},
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
