package common

import (
	"context"
	"database/sql"
	"errors"
	"image"
	"net/url"
	"path"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestCreateCommentValidatesGalleryImages(t *testing.T) {
	imageID := "abcd1234.jpg"
	imagePath := path.Join("/", imageID[:2], imageID[2:4], imageID)
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

	t.Run("accepts cached image without gallery validation", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.CreateCommentInSectionForCommenterResult = 42
		cd := NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
		cacheText := "[img cache:58f3984f584548271144122c7d139e9a6a8a1ad3.png]"
		if _, err := cd.CreateCommentInSectionForCommenter("forum", "topic", 1, 1, 9, 1, cacheText); err != nil {
			t.Fatalf("expected cache image acceptance: %v", err)
		}
		if len(queries.ListUploadedImagePathsByUserCalls) != 0 {
			t.Fatalf("expected no gallery lookup, got %d calls", len(queries.ListUploadedImagePathsByUserCalls))
		}
		if len(queries.ListThreadImagePathsCalls) != 0 {
			t.Fatalf("expected no thread image lookup, got %d calls", len(queries.ListThreadImagePathsCalls))
		}
		if len(queries.CreateThreadImageCalls) != 0 {
			t.Fatalf("expected no thread image record, got %d calls", len(queries.CreateThreadImageCalls))
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
		_, err := cd.CreateCommentInSectionForCommenter("forum", "topic", 1, 1, 9, 1, text)
		if err == nil {
			t.Fatal("expected error for missing gallery image")
		}
		var userError UserError
		if !errors.As(err, &userError) {
			t.Fatalf("expected user error, got %T: %v", err, err)
		}
		if userError.UserErrorMessage() != "One or more images are unavailable. Please upload them again." {
			t.Fatalf("user error = %q", userError.UserErrorMessage())
		}
		if len(queries.CreateCommentInSectionForCommenterCalls) != 0 {
			t.Fatalf("expected no comment creation, got %d calls", len(queries.CreateCommentInSectionForCommenterCalls))
		}
	})

	t.Run("rejects invalid image ref", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		cd := NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
		invalidText := "[img invalid]"
		_, err := cd.CreateCommentInSectionForCommenter("forum", "topic", 1, 1, 9, 1, invalidText)
		if err == nil {
			t.Fatal("expected error for invalid image ref")
		}
		// Expect the error to contain the invalid reference
		if !strings.Contains(err.Error(), "invalid") {
			t.Errorf("error %q should contain 'invalid'", err)
		}
	})

	t.Run("rejects invalid cache image ref", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		cd := NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
		invalidText := "[img cache:../secret.png]"
		_, err := cd.CreateCommentInSectionForCommenter("forum", "topic", 1, 1, 9, 1, invalidText)
		if err == nil {
			t.Fatal("expected error for invalid cache image ref")
		}
		if !strings.Contains(err.Error(), "invalid cache image id") {
			t.Errorf("error %q should contain 'invalid cache image id'", err)
		}
		if len(queries.CreateCommentInSectionForCommenterCalls) != 0 {
			t.Fatalf("expected no comment creation, got %d calls", len(queries.CreateCommentInSectionForCommenterCalls))
		}
	})
}

func TestMapImageURLUsesDefaultThumbnailForLargeUploadedImage(t *testing.T) {
	imageID := "abcd1234.png"
	imagePath := path.Join("/", imageID[:2], imageID[2:4], imageID)
	queries := testhelpers.NewQuerierStub()
	queries.GetUploadedImageByPathReturns = &db.UploadedImage{
		Iduploadedimage: 42,
		Path:            sql.NullString{String: imagePath, Valid: true},
		Width:           sql.NullInt32{Int32: 640, Valid: true},
		Height:          sql.NullInt32{Int32: 480, Valid: true},
	}
	cfg := &config.RuntimeConfig{BaseURL: "https://example.test", ImageThumbnailSizes: "128x256,256x512"}
	cd := NewCoreData(context.Background(), queries, cfg, WithImageSignKey("test-key"))

	mapped := cd.MapImageURL("img", "image:"+imageID)
	parsed, err := url.Parse(mapped)
	if err != nil {
		t.Fatalf("parse mapped URL: %v", err)
	}
	if parsed.Path != "/images/cache/abcd1234_thumb_128x256.png" {
		t.Fatalf("mapped path = %q", parsed.Path)
	}
	if len(queries.GetUploadedImageByPathCalls) != 1 || queries.GetUploadedImageByPathCalls[0].String != imagePath {
		t.Fatalf("uploaded image lookup = %#v", queries.GetUploadedImageByPathCalls)
	}
}

func TestMapImageURLUsesDefaultThumbnailForLargeCachedImage(t *testing.T) {
	imageID := "abcd1234.jpg"
	queries := testhelpers.NewQuerierStub()
	queries.GetImageCacheEntryFn = func(ctx context.Context, id string) (*db.ImageCacheEntry, error) {
		if id != imageID {
			return nil, nil
		}
		return &db.ImageCacheEntry{
			ID:     imageID,
			Width:  sql.NullInt32{Int32: 1600, Valid: true},
			Height: sql.NullInt32{Int32: 1200, Valid: true},
		}, nil
	}
	cfg := &config.RuntimeConfig{BaseURL: "https://example.test", ImageThumbnailSizes: "400x800"}
	cd := NewCoreData(context.Background(), queries, cfg, WithImageSignKey("test-key"))

	mapped := cd.MapImageURL("img", "cache:"+imageID)
	parsed, err := url.Parse(mapped)
	if err != nil {
		t.Fatalf("parse mapped URL: %v", err)
	}
	if parsed.Path != "/images/cache/abcd1234_thumb_400x800.jpg" {
		t.Fatalf("mapped path = %q", parsed.Path)
	}

	full := cd.MapFullImageURL("img", "cache:"+imageID)
	fullParsed, err := url.Parse(full)
	if err != nil {
		t.Fatalf("parse full URL: %v", err)
	}
	if fullParsed.Path != "/images/cache/"+imageID {
		t.Fatalf("full-size path = %q", fullParsed.Path)
	}
}

func TestMapImageURLUsesThumbnailForCachedImageWithoutMetadata(t *testing.T) {
	imageID := "abcd1234.png"
	queries := testhelpers.NewQuerierStub()
	queries.GetImageCacheEntryReturns = nil
	cfg := &config.RuntimeConfig{BaseURL: "https://example.test", ImageThumbnailSizes: "400x800"}
	cd := NewCoreData(context.Background(), queries, cfg, WithImageSignKey("test-key"))

	mapped := cd.MapImageURL("img", "cache:"+imageID)
	parsed, err := url.Parse(mapped)
	if err != nil {
		t.Fatalf("parse mapped URL: %v", err)
	}
	if parsed.Path != "/images/cache/abcd1234_thumb_400x800.png" {
		t.Fatalf("mapped path = %q", parsed.Path)
	}
}

func TestRecordUploadedImageThumbnailLinksSourceImage(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	cd := NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	source := &db.UploadedImage{
		Iduploadedimage: 42,
		Path:            sql.NullString{String: "/ab/cd/abcd1234.png", Valid: true},
	}
	if err := cd.RecordUploadedImageThumbnail(context.Background(), "abcd1234_thumb_128x256.png", source, []byte("thumbnail"), 128, 256); err != nil {
		t.Fatalf("RecordUploadedImageThumbnail: %v", err)
	}
	if len(queries.UpsertImageCacheEntryCalls) != 1 {
		t.Fatalf("cache entry calls = %d", len(queries.UpsertImageCacheEntryCalls))
	}
	entry := queries.UpsertImageCacheEntryCalls[0]
	if !entry.UploadedImageID.Valid || entry.UploadedImageID.Int32 != source.Iduploadedimage {
		t.Fatalf("uploaded image back reference = %#v", entry.UploadedImageID)
	}
	if entry.ID != "abcd1234_thumb_128x256.png" || entry.SourceKind != imageCacheSourceKindUploaded {
		t.Fatalf("cache entry = %#v", entry)
	}
}

func TestStoreImageRecordsDefaultThumbnail(t *testing.T) {
	provider := newMemoryCacheProvider()
	providerName := registerMemoryCacheProvider(t, provider)
	queries := testhelpers.NewQuerierStub(testhelpers.WithGrant("images", "upload", "post"))
	queries.CreateUploadedImageForUploaderResult = 42
	cfg := config.NewRuntimeConfig()
	cfg.ImageUploadProvider = providerName
	cfg.ImageCacheProvider = providerName
	cfg.ImageThumbnailSizes = "64x128,128x256"
	cd := NewCoreData(context.Background(), queries, cfg)
	cd.UserID = 1

	imageID := "abcd1234"
	if _, err := cd.StoreImage(StoreImageParams{
		ID:         imageID,
		Ext:        ".png",
		Data:       []byte("image"),
		Image:      image.NewRGBA(image.Rect(0, 0, 640, 480)),
		UploaderID: 1,
	}); err != nil {
		t.Fatalf("StoreImage: %v", err)
	}
	thumbnailID := "abcd1234_thumb_64x128.png"
	key := path.Join(imageID[:2], imageID[2:4], thumbnailID)
	if _, err := provider.Read(context.Background(), key); err != nil {
		t.Fatalf("read default thumbnail: %v", err)
	}
	if len(queries.UpsertImageCacheEntryCalls) != 1 {
		t.Fatalf("cache entry calls = %d", len(queries.UpsertImageCacheEntryCalls))
	}
	entry := queries.UpsertImageCacheEntryCalls[0]
	if entry.ID != thumbnailID || !entry.UploadedImageID.Valid || entry.UploadedImageID.Int32 != 42 {
		t.Fatalf("cache entry = %#v", entry)
	}
}

func TestSanitizeCodeImagesQueuesImageAliasGoogleRedirect(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	cfg := config.NewRuntimeConfig()
	cd := NewCoreData(context.Background(), queries, cfg)
	target := "https://www.pinterest.com/pin/88383211424382477/"
	googleURL := "https://www.google.com/url?sa=t&source=web&rct=j&url=" + url.QueryEscape(target) + "&ved=0CBYQjRxqFwoTCODsyOfE1pUDFQAAAAAdAAAAABBq&opi=89978449"

	got, queued := cd.sanitizeCodeImagesAndQueue("[image " + googleURL + "]")
	if !strings.Contains(got, "[img=cache:") {
		t.Fatalf("sanitized code = %q, want cache image", got)
	}
	if len(queued) != 1 {
		t.Fatalf("queued fetches = %d, want 1", len(queued))
	}
	if queued[0].sourceURL != target {
		t.Fatalf("queued source url = %q, want %q", queued[0].sourceURL, target)
	}
	if len(queries.CreatePendingImageCacheEntryCalls) != 1 {
		t.Fatalf("pending cache calls = %d, want 1", len(queries.CreatePendingImageCacheEntryCalls))
	}
	call := queries.CreatePendingImageCacheEntryCalls[0]
	if !call.SourceUrl.Valid || call.SourceUrl.String != target {
		t.Fatalf("pending source url = %#v, want %q", call.SourceUrl, target)
	}
}
