package common

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"image"
	_ "image/gif"  // Register format
	_ "image/jpeg" // Register format
	_ "image/png"  // Register format
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	intimages "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/opengraph"
	"github.com/arran4/goa4web/internal/upload"
)

const (
	// imageCacheSourceKindRemote identifies cache entries fetched from an external URL.
	imageCacheSourceKindRemote = "remote"
	// imageCacheSourceKindUploaded identifies cache entries derived from an uploaded image.
	imageCacheSourceKindUploaded = "uploaded"
	// imageCacheSourceKindDerived identifies cache entries derived from another cached image.
	imageCacheSourceKindDerived = "derived"
	// imageCacheStatusReady identifies cache entries with materialized content.
	imageCacheStatusReady = "ready"
	// imageCacheStatusPending identifies cache entries awaiting download.
	imageCacheStatusPending = "pending"
)

type downloadedImage struct {
	body             []byte
	contentType      string
	contentExpiresAt sql.NullTime
	width            sql.NullInt32
	height           sql.NullInt32
	checksum         string
	thumbnailID      sql.NullString
	thumbnailBytes   []byte
	thumbnailHeight  int
	thumbnailWidth   int
}

// DownloadAndCacheImage downloads an image from a URL, stores it in the image
// cache, and returns the stored cache ID prefixed with "cache:".
func (cd *CoreData) DownloadAndCacheImage(imgURL string) (string, error) {
	img, err := cd.downloadExternalImage(imgURL)
	if err != nil {
		return "", err
	}

	hashBytes := sha256.Sum256(img.body)
	hash := fmt.Sprintf("%x", hashBytes[:20])
	// Try to get extension from URL path
	u, err := url.Parse(imgURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}
	ext, err := intimages.CleanExtension(path.Base(u.Path))
	if err != nil {
		// Fallback to .jpg if no valid extension found in URL
		ext = ".jpg"
	}

	cacheRef := hash + ext
	sub1, sub2 := hash[:2], hash[2:4]
	key := path.Join(sub1, sub2, cacheRef)

	if err := cd.writeImageCacheBytes(context.Background(), key, cacheRef, img.body); err != nil {
		return "", err
	}

	if err := cd.writeRemoteImageThumbnail(context.Background(), cacheRef, ext, img); err != nil {
		return "", err
	}

	if err := cd.recordRemoteImageCacheEntry(context.Background(), cacheRef, imgURL, img, time.Now().UTC()); err != nil {
		return "", err
	}

	return "cache:" + cacheRef, nil
}

// QueueRemoteImageCache records a pending remote image cache entry and returns
// its stable cache reference. A background worker can later materialize it.
func (cd *CoreData) QueueRemoteImageCache(imgURL string) (string, error) {
	imgURL = canonicalRemoteImageSourceURL(imgURL)
	if cd == nil || cd.queries == nil {
		return cd.DownloadAndCacheImage(imgURL)
	}
	id, err := remoteImageCacheIDForURL(imgURL)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	if err := cd.queries.CreatePendingImageCacheEntry(cd.ctx, db.CreatePendingImageCacheEntryParams{
		ID:            id,
		SourceUrl:     sql.NullString{String: imgURL, Valid: imgURL != ""},
		SourceKind:    imageCacheSourceKindRemote,
		CreatedAt:     now,
		LastUsedAt:    sql.NullTime{Time: now, Valid: true},
		NextAttemptAt: sql.NullTime{Time: now, Valid: true},
	}); err != nil {
		return "", err
	}
	return "cache:" + id, nil
}

// StartRemoteImageCacheFetch starts a best-effort immediate fetch for a queued
// remote cache entry. The scheduled worker remains responsible for retries.
func (cd *CoreData) StartRemoteImageCacheFetch(id, sourceURL string) {
	if cd == nil || cd.queries == nil || id == "" || sourceURL == "" {
		return
	}
	sourceURL = canonicalRemoteImageSourceURL(sourceURL)
	go func() {
		ctx := context.Background()
		attemptTime := time.Now().UTC()
		log.Printf("image cache fetch start: mode=immediate id=%s source=%q attempt_at=%s", id, sourceURL, attemptTime.Format(time.RFC3339))
		if err := cd.refreshRemoteImageCacheEntry(ctx, id, sourceURL, attemptTime); err != nil {
			log.Printf("image cache fetch failed: mode=immediate id=%s source=%q error=%q", id, sourceURL, err.Error())
			retryDelay := cd.imageCacheFetchRetryDelay()
			maxRetries := cd.imageCacheFetchMaxRetries()
			if recErr := cd.queries.RecordImageCacheFetchFailure(ctx, db.RecordImageCacheFetchFailureParams{
				ID:            id,
				RetryCount:    int32(maxRetries),
				RetryCount_2:  int32(maxRetries),
				ErrorMessage:  sql.NullString{String: err.Error(), Valid: true},
				LastAttemptAt: sql.NullTime{Time: attemptTime, Valid: true},
				NextAttemptAt: sql.NullTime{Time: attemptTime.Add(retryDelay), Valid: true},
			}); recErr != nil {
				log.Printf("record immediate image cache fetch failure: %v", recErr)
			}
			return
		}
		log.Printf("image cache fetch complete: mode=immediate id=%s source=%q", id, sourceURL)
	}()
}

// ProcessPendingRemoteImageCacheEntries materializes queued remote cache
// entries. Failed downloads are marked failed so requests keep seeing a
// placeholder instead of blocking repeatedly.
func (cd *CoreData) ProcessPendingRemoteImageCacheEntries(ctx context.Context, limit int32) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	if limit <= 0 {
		limit = 10
	}
	now := time.Now().UTC()
	maxRetries := cd.imageCacheFetchMaxRetries()
	entries, err := cd.queries.ListDuePendingImageCacheEntries(ctx, db.ListDuePendingImageCacheEntriesParams{
		RetryCount:    int32(maxRetries),
		NextAttemptAt: sql.NullTime{Time: now, Valid: true},
		Limit:         limit,
	})
	if err != nil {
		return fmt.Errorf("list pending image cache entries: %w", err)
	}
	for _, entry := range entries {
		if entry == nil || !entry.SourceUrl.Valid || entry.SourceUrl.String == "" {
			continue
		}
		attemptTime := time.Now().UTC()
		log.Printf("image cache fetch start: mode=worker id=%s source=%q attempt_at=%s retry_count=%d", entry.ID, entry.SourceUrl.String, attemptTime.Format(time.RFC3339), entry.RetryCount)
		if err := cd.refreshRemoteImageCacheEntry(ctx, entry.ID, entry.SourceUrl.String, attemptTime); err != nil {
			log.Printf("image cache fetch failed: mode=worker id=%s source=%q retry_count=%d error=%q", entry.ID, entry.SourceUrl.String, entry.RetryCount, err.Error())
			retryDelay := cd.imageCacheFetchRetryDelay()
			if recErr := cd.queries.RecordImageCacheFetchFailure(ctx, db.RecordImageCacheFetchFailureParams{
				ID:            entry.ID,
				RetryCount:    int32(maxRetries),
				RetryCount_2:  int32(maxRetries),
				ErrorMessage:  sql.NullString{String: err.Error(), Valid: true},
				LastAttemptAt: sql.NullTime{Time: attemptTime, Valid: true},
				NextAttemptAt: sql.NullTime{Time: attemptTime.Add(retryDelay), Valid: true},
			}); recErr != nil {
				log.Printf("record worker image cache fetch failure: id=%s source=%q error=%q", entry.ID, entry.SourceUrl.String, recErr.Error())
			}
			continue
		}
		log.Printf("image cache fetch complete: mode=worker id=%s source=%q", entry.ID, entry.SourceUrl.String)
	}
	return nil
}

func remoteImageCacheIDForURL(imgURL string) (string, error) {
	u, err := url.Parse(imgURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}
	ext, err := intimages.CleanExtension(path.Base(u.Path))
	if err != nil {
		ext = ".jpg"
	}
	hashBytes := sha256.Sum256([]byte(imgURL))
	hash := fmt.Sprintf("%x", hashBytes[:20])
	return hash + ext, nil
}

func canonicalRemoteImageSourceURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || !u.IsAbs() {
		return raw
	}
	host := strings.ToLower(u.Hostname())
	if isGoogleRedirectHost(host) && (u.Path == "/url" || u.Path == "/imgres") {
		for _, key := range []string{"url", "q", "imgurl"} {
			next := u.Query().Get(key)
			if next == "" {
				continue
			}
			nu, err := url.Parse(next)
			if err == nil && nu.IsAbs() && (nu.Scheme == "http" || nu.Scheme == "https") {
				return nu.String()
			}
		}
	}
	return raw
}

func isGoogleRedirectHost(host string) bool {
	return host == "google.com" || host == "www.google.com" || strings.HasSuffix(host, ".google.com")
}

func isImageContentType(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	return strings.HasPrefix(contentType, "image/")
}

func isHTMLContentType(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	return contentType == "text/html" || contentType == "application/xhtml+xml"
}

func looksLikeHTML(body []byte) bool {
	s := strings.TrimSpace(strings.ToLower(string(body[:min(len(body), 512)])))
	return strings.HasPrefix(s, "<!doctype html") || strings.HasPrefix(s, "<html") || strings.Contains(s, "<head")
}

func (cd *CoreData) downloadExternalImage(imgURL string) (*downloadedImage, error) {
	return cd.downloadExternalImageAtDepth(canonicalRemoteImageSourceURL(imgURL), 0)
}

func (cd *CoreData) downloadExternalImageAtDepth(imgURL string, depth int) (*downloadedImage, error) {
	if depth > 3 {
		return nil, fmt.Errorf("too many image metadata redirects")
	}
	client := opengraph.NewSafeClient() // Always use a safe client for external URLs
	if cd != nil && cd.HTTPClient() != nil {
		client = cd.HTTPClient()
	}

	resp, err := client.Get(imgURL)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()
	log.Printf("image cache fetch http response: url=%q status=%q content_type=%q", imgURL, resp.Status, resp.Header.Get("Content-Type"))
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("http status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("empty body")
	}
	log.Printf("image cache fetch http body: url=%q bytes=%d", imgURL, len(body))
	hashBytes := sha256.Sum256(body)
	contentType := resp.Header.Get("Content-Type")
	img := &downloadedImage{
		body:             body,
		contentType:      contentType,
		contentExpiresAt: httpContentExpiresAt(resp.Header, time.Now().UTC()),
		checksum:         fmt.Sprintf("%x", hashBytes[:]),
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(body))
	if err == nil {
		img.width = sql.NullInt32{Int32: int32(cfg.Width), Valid: cfg.Width > 0}
		img.height = sql.NullInt32{Int32: int32(cfg.Height), Valid: cfg.Height > 0}
		return img, nil
	}
	if isImageContentType(contentType) {
		return img, nil
	}
	if !isHTMLContentType(contentType) && !looksLikeHTML(body) {
		return nil, fmt.Errorf("not an image: %s", contentType)
	}
	info, err := opengraph.Fetch(imgURL, client)
	if err != nil {
		return nil, fmt.Errorf("fetch image metadata: %w", err)
	}
	if info == nil || info.Image == "" {
		return nil, fmt.Errorf("html document did not expose an image")
	}
	nextURL, err := url.Parse(info.Image)
	if err != nil {
		return nil, fmt.Errorf("parse metadata image url: %w", err)
	}
	baseURL, err := url.Parse(imgURL)
	if err != nil {
		return nil, fmt.Errorf("parse base image url: %w", err)
	}
	resolved := baseURL.ResolveReference(nextURL).String()
	if resolved == imgURL {
		return nil, fmt.Errorf("metadata image points to same url")
	}
	log.Printf("image cache fetch metadata image: page=%q image=%q", imgURL, resolved)
	return cd.downloadExternalImageAtDepth(resolved, depth+1)
}

func (cd *CoreData) writeImageCacheBytes(ctx context.Context, key, id string, body []byte) error {
	if cp := upload.CacheProviderFromConfig(cd.Config); cp != nil {
		if err := cp.Write(ctx, key, body); err != nil {
			return fmt.Errorf("cache write: %w", err)
		}
		return nil
	}

	fullPath := path.Join(cd.Config.ImageCacheDir, path.Dir(key), id)
	if err := os.MkdirAll(path.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("mkdir cache: %w", err)
	}
	if err := os.WriteFile(fullPath, body, 0644); err != nil {
		return fmt.Errorf("write cache file: %w", err)
	}
	return nil
}

func (cd *CoreData) writeRemoteImageThumbnail(ctx context.Context, id, ext string, img *downloadedImage) error {
	if img == nil || len(img.body) == 0 {
		return nil
	}
	src, _, err := image.Decode(bytes.NewReader(img.body))
	if err != nil {
		return nil
	}
	generator := "bild"
	size := config.ThumbnailSize{Width: config.DefaultImageThumbnailWidth, Height: config.DefaultImageThumbnailHeight}
	if cd != nil && cd.Config != nil {
		if cd.Config.ImageThumbnailGenerator != "" {
			generator = cd.Config.ImageThumbnailGenerator
		}
		size = cd.Config.ThumbnailSizes()[0]
	}
	thumbBytes, err := intimages.GenerateThumbnailWithinBounds(src, ext, generator, size.Height, size.Width)
	if err != nil {
		return fmt.Errorf("generate cache thumbnail: %w", err)
	}
	thumbID := thumbnailFilename(strings.TrimSuffix(id, ext), ext, size)
	key, err := imageCacheKey(thumbID)
	if err != nil {
		return err
	}
	if err := cd.writeImageCacheBytes(ctx, key, thumbID, thumbBytes); err != nil {
		return err
	}
	img.thumbnailID = sql.NullString{String: thumbID, Valid: true}
	img.thumbnailBytes = thumbBytes
	img.thumbnailHeight, img.thumbnailWidth, err = intimages.DimensionsWithinBounds(src, size.Height, size.Width)
	if err != nil {
		return fmt.Errorf("cache thumbnail dimensions: %w", err)
	}
	return nil
}

func (cd *CoreData) recordRemoteImageCacheEntry(ctx context.Context, id, sourceURL string, img *downloadedImage, now time.Time) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	if err := cd.queries.UpsertImageCacheEntry(ctx, db.UpsertImageCacheEntryParams{
		ID:               id,
		SourceUrl:        sql.NullString{String: sourceURL, Valid: sourceURL != ""},
		SourceKind:       imageCacheSourceKindRemote,
		Status:           imageCacheStatusReady,
		CreatedAt:        now,
		LastUsedAt:       sql.NullTime{Time: now, Valid: true},
		FetchedAt:        sql.NullTime{Time: now, Valid: true},
		ExpiresAt:        cd.remoteImageCacheExpiresAt(now, img.contentExpiresAt),
		ContentExpiresAt: img.contentExpiresAt,
		ContentType:      sql.NullString{String: img.contentType, Valid: img.contentType != ""},
		SizeBytes:        sql.NullInt64{Int64: int64(len(img.body)), Valid: len(img.body) > 0},
		Width:            img.width,
		Height:           img.height,
		Checksum:         sql.NullString{String: img.checksum, Valid: img.checksum != ""},
		ThumbnailID:      img.thumbnailID,
	}); err != nil {
		return err
	}
	if !img.thumbnailID.Valid || len(img.thumbnailBytes) == 0 {
		return nil
	}
	return cd.queries.UpsertImageCacheEntry(ctx, db.UpsertImageCacheEntryParams{
		ID:          img.thumbnailID.String,
		SourceUrl:   sql.NullString{String: sourceURL, Valid: sourceURL != ""},
		SourceKind:  imageCacheSourceKindRemote,
		Status:      imageCacheStatusReady,
		CreatedAt:   now,
		LastUsedAt:  sql.NullTime{Time: now, Valid: true},
		FetchedAt:   sql.NullTime{Time: now, Valid: true},
		ContentType: sql.NullString{String: mime.TypeByExtension(filepath.Ext(img.thumbnailID.String)), Valid: true},
		SizeBytes:   sql.NullInt64{Int64: int64(len(img.thumbnailBytes)), Valid: true},
		Width:       sql.NullInt32{Int32: int32(img.thumbnailWidth), Valid: img.thumbnailWidth > 0},
		Height:      sql.NullInt32{Int32: int32(img.thumbnailHeight), Valid: img.thumbnailHeight > 0},
	})
}

// RecordUploadedImageThumbnail records a generated thumbnail and its uploaded-image source.
func (cd *CoreData) RecordUploadedImageThumbnail(ctx context.Context, thumbnailID string, source *db.UploadedImage, body []byte, height, width int) error {
	return cd.recordUploadedImageCacheEntry(ctx, thumbnailID, source, body, height, width, imageCacheSourceKindUploaded)
}

// RecordUploadedImageDerivative records a non-thumbnail cache derivative of an uploaded image.
func (cd *CoreData) RecordUploadedImageDerivative(ctx context.Context, id string, source *db.UploadedImage, body []byte, height, width int) error {
	return cd.recordUploadedImageCacheEntry(ctx, id, source, body, height, width, imageCacheSourceKindDerived)
}

// RecordCachedImageThumbnail records a generated thumbnail for a cached image.
func (cd *CoreData) RecordCachedImageThumbnail(ctx context.Context, thumbnailID string, source *db.ImageCacheEntry, body []byte, height, width int) error {
	if source == nil {
		return cd.RecordDerivedImageCacheEntry(ctx, thumbnailID, body)
	}
	if cd == nil || cd.queries == nil {
		return nil
	}
	sourceKind := source.SourceKind
	if sourceKind == "" {
		sourceKind = imageCacheSourceKindDerived
	}
	now := time.Now().UTC()
	return cd.queries.UpsertImageCacheEntry(ctx, db.UpsertImageCacheEntryParams{
		ID:          thumbnailID,
		SourceUrl:   source.SourceUrl,
		SourceKind:  sourceKind,
		Status:      imageCacheStatusReady,
		CreatedAt:   now,
		LastUsedAt:  sql.NullTime{Time: now, Valid: true},
		FetchedAt:   sql.NullTime{Time: now, Valid: true},
		ContentType: sql.NullString{String: mime.TypeByExtension(filepath.Ext(thumbnailID)), Valid: true},
		SizeBytes:   sql.NullInt64{Int64: int64(len(body)), Valid: true},
		Width:       sql.NullInt32{Int32: int32(width), Valid: width > 0},
		Height:      sql.NullInt32{Int32: int32(height), Valid: height > 0},
	})
}

func (cd *CoreData) recordUploadedImageCacheEntry(ctx context.Context, id string, source *db.UploadedImage, body []byte, height, width int, sourceKind string) error {
	if cd == nil || cd.queries == nil || source == nil || source.Iduploadedimage == 0 {
		return nil
	}
	now := time.Now().UTC()
	return cd.queries.UpsertImageCacheEntry(ctx, db.UpsertImageCacheEntryParams{
		ID:              id,
		SourceUrl:       source.Path,
		SourceKind:      sourceKind,
		Status:          imageCacheStatusReady,
		CreatedAt:       now,
		LastUsedAt:      sql.NullTime{Time: now, Valid: true},
		FetchedAt:       sql.NullTime{Time: now, Valid: true},
		ContentType:     sql.NullString{String: mime.TypeByExtension(filepath.Ext(id)), Valid: true},
		SizeBytes:       sql.NullInt64{Int64: int64(len(body)), Valid: true},
		Width:           sql.NullInt32{Int32: int32(width), Valid: width > 0},
		Height:          sql.NullInt32{Int32: int32(height), Valid: height > 0},
		UploadedImageID: sql.NullInt32{Int32: source.Iduploadedimage, Valid: true},
	})
}

// RecordDerivedImageCacheEntry records a non-thumbnail derivative stored in the image cache.
func (cd *CoreData) RecordDerivedImageCacheEntry(ctx context.Context, id string, body []byte) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	now := time.Now().UTC()
	return cd.queries.UpsertImageCacheEntry(ctx, db.UpsertImageCacheEntryParams{
		ID:          id,
		SourceKind:  imageCacheSourceKindDerived,
		Status:      imageCacheStatusReady,
		CreatedAt:   now,
		LastUsedAt:  sql.NullTime{Time: now, Valid: true},
		FetchedAt:   sql.NullTime{Time: now, Valid: true},
		ContentType: sql.NullString{String: mime.TypeByExtension(filepath.Ext(id)), Valid: true},
		SizeBytes:   sql.NullInt64{Int64: int64(len(body)), Valid: true},
	})
}

func (cd *CoreData) remoteImageCacheExpiresAt(now time.Time, contentExpiresAt sql.NullTime) sql.NullTime {
	mode := cd.imageCacheMode()
	switch mode {
	case "http", "http-size":
		return contentExpiresAt
	case "ttl", "ttl-size":
		ttl := cd.imageCacheTTL()
		if ttl <= 0 {
			return sql.NullTime{}
		}
		return sql.NullTime{Time: now.Add(ttl), Valid: true}
	default:
		return sql.NullTime{}
	}
}

func (cd *CoreData) imageCacheTTL() time.Duration {
	if cd == nil || cd.Config == nil || cd.Config.ImageCacheTTL == "" {
		return 0
	}
	d, err := parseCacheDuration(cd.Config.ImageCacheTTL)
	if err != nil {
		return 0
	}
	return d
}

func (cd *CoreData) imageCacheFetchMaxRetries() int {
	if cd == nil || cd.Config == nil || cd.Config.ImageCacheFetchMaxRetries <= 0 {
		return 3
	}
	return cd.Config.ImageCacheFetchMaxRetries
}

func (cd *CoreData) imageCacheFetchRetryDelay() time.Duration {
	if cd == nil || cd.Config == nil || cd.Config.ImageCacheFetchRetryDelay == "" {
		return time.Minute
	}
	d, err := parseCacheDuration(cd.Config.ImageCacheFetchRetryDelay)
	if err != nil || d <= 0 {
		return time.Minute
	}
	return d
}

func (cd *CoreData) imageCacheMode() string {
	if cd == nil || cd.Config == nil || cd.Config.ImageCacheMode == "" {
		return "none"
	}
	return strings.ToLower(strings.TrimSpace(cd.Config.ImageCacheMode))
}

func imageCacheKey(id string) (string, error) {
	if !intimages.ValidID(id) {
		return "", fmt.Errorf("invalid cache id")
	}
	return path.Join(id[:2], id[2:4], id), nil
}

// PrepareImageCacheEntryForServe refreshes an expired remote cache entry when
// external cache expiry is enabled. It returns false when the cached object
// should not be served.
func (cd *CoreData) PrepareImageCacheEntryForServe(ctx context.Context, id string) (bool, error) {
	if cd == nil || cd.queries == nil {
		return true, nil
	}
	entry, err := cd.queries.GetImageCacheEntry(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		return false, fmt.Errorf("get image cache entry: %w", err)
	}
	if entry == nil {
		return true, nil
	}
	now := time.Now().UTC()
	if entry.SourceKind != imageCacheSourceKindRemote {
		_ = cd.queries.TouchImageCacheEntry(ctx, db.TouchImageCacheEntryParams{ID: id, LastUsedAt: sql.NullTime{Time: now, Valid: true}})
		return true, nil
	}
	if entry.Status != imageCacheStatusReady {
		return false, nil
	}
	if cd.imageCacheMode() == "none" {
		_ = cd.queries.TouchImageCacheEntry(ctx, db.TouchImageCacheEntryParams{ID: id, LastUsedAt: sql.NullTime{Time: now, Valid: true}})
		return true, nil
	}
	if !cd.imageCacheEntryExpired(entry, now) {
		_ = cd.queries.TouchImageCacheEntry(ctx, db.TouchImageCacheEntryParams{ID: id, LastUsedAt: sql.NullTime{Time: now, Valid: true}})
		return true, nil
	}
	if !entry.SourceUrl.Valid || entry.SourceUrl.String == "" {
		return false, nil
	}
	if err := cd.refreshRemoteImageCacheEntry(ctx, id, entry.SourceUrl.String, now); err != nil {
		return false, err
	}
	return true, nil
}

func (cd *CoreData) imageCacheEntryExpired(entry *db.ImageCacheEntry, now time.Time) bool {
	if entry == nil {
		return false
	}
	switch cd.imageCacheMode() {
	case "last-used", "last-used-size":
		ttl := cd.imageCacheTTL()
		if ttl <= 0 {
			return false
		}
		lastUsed := entry.CreatedAt
		if entry.LastUsedAt.Valid {
			lastUsed = entry.LastUsedAt.Time
		}
		return !now.Before(lastUsed.Add(ttl))
	case "ttl", "ttl-size", "http", "http-size":
		return entry.ExpiresAt.Valid && !now.Before(entry.ExpiresAt.Time)
	default:
		return false
	}
}

func (cd *CoreData) refreshRemoteImageCacheEntry(ctx context.Context, id, sourceURL string, now time.Time) error {
	img, err := cd.downloadExternalImage(sourceURL)
	if err != nil {
		return err
	}
	key, err := imageCacheKey(id)
	if err != nil {
		return err
	}
	if err := cd.writeImageCacheBytes(ctx, key, id, img.body); err != nil {
		return err
	}
	ext := path.Ext(id)
	if err := cd.writeRemoteImageThumbnail(ctx, id, ext, img); err != nil {
		return err
	}
	return cd.recordRemoteImageCacheEntry(ctx, id, sourceURL, img, now)
}

// ImageCacheEntry returns metadata for a cache entry when present.
func (cd *CoreData) ImageCacheEntry(ctx context.Context, id string) (*db.ImageCacheEntry, error) {
	if cd == nil || cd.queries == nil {
		return nil, sql.ErrNoRows
	}
	return cd.queries.GetImageCacheEntry(ctx, id)
}

func parseCacheDuration(raw string) (time.Duration, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if d, err := time.ParseDuration(raw); err == nil {
		return d, nil
	}
	multiplier := time.Duration(0)
	switch {
	case strings.HasSuffix(raw, "d"):
		multiplier = 24 * time.Hour
	case strings.HasSuffix(raw, "w"):
		multiplier = 7 * 24 * time.Hour
	case strings.HasSuffix(raw, "y"):
		multiplier = 365 * 24 * time.Hour
	default:
		return 0, fmt.Errorf("invalid duration")
	}
	n, err := strconv.ParseFloat(strings.TrimSpace(raw[:len(raw)-1]), 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(n * float64(multiplier)), nil
}

func httpContentExpiresAt(h http.Header, now time.Time) sql.NullTime {
	if cc := strings.ToLower(h.Get("Cache-Control")); cc != "" {
		for _, part := range strings.Split(cc, ",") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "max-age=") {
				secs, err := strconv.ParseInt(strings.TrimPrefix(part, "max-age="), 10, 64)
				if err == nil && secs >= 0 {
					return sql.NullTime{Time: now.Add(time.Duration(secs) * time.Second), Valid: true}
				}
			}
		}
	}
	if expires := h.Get("Expires"); expires != "" {
		if t, err := http.ParseTime(expires); err == nil {
			return sql.NullTime{Time: t.UTC(), Valid: true}
		}
	}
	return sql.NullTime{}
}
