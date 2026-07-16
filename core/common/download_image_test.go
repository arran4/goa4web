package common

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/arran4/goa4web/internal/upload"
)

type memoryCacheProvider struct {
	mu    sync.Mutex
	files map[string][]byte
}

func newMemoryCacheProvider() *memoryCacheProvider {
	return &memoryCacheProvider{files: map[string][]byte{}}
}

func (p *memoryCacheProvider) Check(context.Context) error { return nil }

func (p *memoryCacheProvider) Write(_ context.Context, name string, data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.files[name] = append([]byte(nil), data...)
	return nil
}

func (p *memoryCacheProvider) Read(_ context.Context, name string) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	data, ok := p.files[name]
	if !ok {
		return nil, fmt.Errorf("missing file")
	}
	return append([]byte(nil), data...), nil
}

func registerMemoryCacheProvider(t *testing.T, provider *memoryCacheProvider) string {
	t.Helper()
	name := "test-cache-" + strings.ReplaceAll(t.Name(), "/", "-")
	upload.RegisterProvider(name, func(*config.RuntimeConfig) upload.Provider { return provider })
	return name
}

func TestDownloadAndCacheImageRecordsRemoteMetadata(t *testing.T) {
	provider := newMemoryCacheProvider()
	providerName := registerMemoryCacheProvider(t, provider)
	queries := testhelpers.NewQuerierStub()
	cfg := config.NewRuntimeConfig()
	cfg.ImageCacheProvider = providerName
	cfg.ImageCacheMode = "ttl"
	cfg.ImageCacheTTL = "2h"
	cd := NewCoreData(context.Background(), queries, cfg)
	var imageBytes bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 3, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 3; x++ {
			img.Set(x, y, color.RGBA{R: 0x20, G: 0x40, B: 0x80, A: 0xff})
		}
	}
	if err := png.Encode(&imageBytes, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	expires := time.Now().UTC().Add(time.Hour).Truncate(time.Second)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Expires", expires.Format(http.TimeFormat))
		_, _ = w.Write(imageBytes.Bytes())
	}))
	defer srv.Close()

	sourceURL := srv.URL + "/logo.png"
	ref, err := cd.DownloadAndCacheImage(sourceURL)
	if err != nil {
		t.Fatalf("DownloadAndCacheImage: %v", err)
	}
	id := strings.TrimPrefix(ref, "cache:")
	key := path.Join(id[:2], id[2:4], id)
	if _, err := provider.Read(context.Background(), key); err != nil {
		t.Fatalf("expected cache write at %s: %v", key, err)
	}
	if len(queries.UpsertImageCacheEntryCalls) != 1 {
		t.Fatalf("expected one metadata upsert, got %d", len(queries.UpsertImageCacheEntryCalls))
	}
	got := queries.UpsertImageCacheEntryCalls[0]
	if got.ID != id {
		t.Fatalf("metadata id = %q, want %q", got.ID, id)
	}
	if !got.SourceUrl.Valid || got.SourceUrl.String != sourceURL {
		t.Fatalf("source url = %#v, want %q", got.SourceUrl, sourceURL)
	}
	if got.SourceKind != imageCacheSourceKindRemote {
		t.Fatalf("source kind = %q, want %q", got.SourceKind, imageCacheSourceKindRemote)
	}
	if got.Status != imageCacheStatusReady {
		t.Fatalf("status = %q, want %q", got.Status, imageCacheStatusReady)
	}
	if !got.ExpiresAt.Valid {
		t.Fatal("expected expiry when ttl cache mode is enabled")
	}
	if got.ExpiresAt.Time.Before(got.CreatedAt.Add(2*time.Hour - time.Minute)) {
		t.Fatalf("expiry %s too early for created_at %s", got.ExpiresAt.Time, got.CreatedAt)
	}
	if !got.ContentType.Valid || got.ContentType.String != "image/png" {
		t.Fatalf("content type = %#v, want image/png", got.ContentType)
	}
	if !got.SizeBytes.Valid {
		t.Fatalf("size bytes = %#v", got.SizeBytes)
	}
	if got.SizeBytes.Int64 != int64(imageBytes.Len()) {
		t.Fatalf("size bytes = %d, want %d", got.SizeBytes.Int64, imageBytes.Len())
	}
	if !got.Checksum.Valid || got.Checksum.String == "" {
		t.Fatalf("expected checksum, got %#v", got.Checksum)
	}
	if !got.Width.Valid || got.Width.Int32 != 3 || !got.Height.Valid || got.Height.Int32 != 2 {
		t.Fatalf("dimensions = %dx%d", got.Width.Int32, got.Height.Int32)
	}
	if !got.ThumbnailID.Valid || got.ThumbnailID.String == "" {
		t.Fatalf("expected thumbnail id, got %#v", got.ThumbnailID)
	}
	if !got.ContentExpiresAt.Valid || !got.ContentExpiresAt.Time.Equal(expires) {
		t.Fatalf("content expiry = %#v, want %s", got.ContentExpiresAt, expires)
	}
}

func TestPrepareImageCacheEntryForServeRefreshesExpiredRemoteEntry(t *testing.T) {
	provider := newMemoryCacheProvider()
	providerName := registerMemoryCacheProvider(t, provider)
	id := "abcd1234.png"
	key := path.Join(id[:2], id[2:4], id)
	if err := provider.Write(context.Background(), key, []byte("stale")); err != nil {
		t.Fatalf("seed cache: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("fresh"))
	}))
	defer srv.Close()

	now := time.Now().UTC()
	queries := testhelpers.NewQuerierStub()
	queries.GetImageCacheEntryReturns = &db.ImageCacheEntry{
		ID:         id,
		SourceUrl:  sql.NullString{String: srv.URL + "/logo.png", Valid: true},
		SourceKind: imageCacheSourceKindRemote,
		Status:     imageCacheStatusReady,
		CreatedAt:  now.Add(-2 * time.Hour),
		ExpiresAt:  sql.NullTime{Time: now.Add(-time.Hour), Valid: true},
	}
	cfg := config.NewRuntimeConfig()
	cfg.ImageCacheProvider = providerName
	cfg.ImageCacheMode = "ttl"
	cfg.ImageCacheTTL = "1h"
	cd := NewCoreData(context.Background(), queries, cfg)

	ok, err := cd.PrepareImageCacheEntryForServe(context.Background(), id)
	if err != nil {
		t.Fatalf("PrepareImageCacheEntryForServe: %v", err)
	}
	if !ok {
		t.Fatal("expected refreshed cache entry to be servable")
	}
	data, err := provider.Read(context.Background(), key)
	if err != nil {
		t.Fatalf("read refreshed cache: %v", err)
	}
	if string(data) != "fresh" {
		t.Fatalf("cache data = %q, want fresh", data)
	}
	if len(queries.UpsertImageCacheEntryCalls) != 1 {
		t.Fatalf("expected metadata refresh upsert, got %d", len(queries.UpsertImageCacheEntryCalls))
	}
	if !queries.UpsertImageCacheEntryCalls[0].ExpiresAt.Valid {
		t.Fatal("expected refreshed entry to have a new expiry")
	}
}

func TestQueueRemoteImageCacheCreatesPendingEntry(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	cfg := config.NewRuntimeConfig()
	cd := NewCoreData(context.Background(), queries, cfg)

	ref, err := cd.QueueRemoteImageCache("https://example.com/logo.png")
	if err != nil {
		t.Fatalf("QueueRemoteImageCache: %v", err)
	}
	if !strings.HasPrefix(ref, "cache:") {
		t.Fatalf("ref = %q, want cache ref", ref)
	}
	if len(queries.CreatePendingImageCacheEntryCalls) != 1 {
		t.Fatalf("expected one pending entry, got %d", len(queries.CreatePendingImageCacheEntryCalls))
	}
	got := queries.CreatePendingImageCacheEntryCalls[0]
	if got.ID != strings.TrimPrefix(ref, "cache:") {
		t.Fatalf("pending id = %q, ref = %q", got.ID, ref)
	}
	if got.SourceKind != imageCacheSourceKindRemote {
		t.Fatalf("source kind = %q", got.SourceKind)
	}
	if !got.SourceUrl.Valid || got.SourceUrl.String != "https://example.com/logo.png" {
		t.Fatalf("source url = %#v", got.SourceUrl)
	}
	if !got.NextAttemptAt.Valid {
		t.Fatal("expected pending entry to have next attempt time")
	}
}

func TestDownloadExternalImageUsesOpenGraphImageFromHTML(t *testing.T) {
	var imageBytes bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 4, 3))
	for y := 0; y < 3; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 0x10, G: 0x20, B: 0x30, A: 0xff})
		}
	}
	if err := png.Encode(&imageBytes, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/pin":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = fmt.Fprintf(w, `<html><head><meta property="og:image" content="%s/photo.png"></head><body></body></html>`, "http://"+r.Host)
		case "/photo.png":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(imageBytes.Bytes())
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cd := NewCoreData(context.Background(), testhelpers.NewQuerierStub(), config.NewRuntimeConfig())
	got, err := cd.downloadExternalImage(srv.URL + "/pin")
	if err != nil {
		t.Fatalf("downloadExternalImage: %v", err)
	}
	if got == nil || !got.width.Valid || got.width.Int32 != 4 || !got.height.Valid || got.height.Int32 != 3 {
		t.Fatalf("dimensions = %#v x %#v, want 4x3", got.width, got.height)
	}
	if got.contentType != "image/png" {
		t.Fatalf("content type = %q, want image/png", got.contentType)
	}
}

func TestCreateCommentStartsImmediateRemoteImageFetch(t *testing.T) {
	var imageBytes bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	if err := png.Encode(&imageBytes, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(imageBytes.Bytes())
	}))
	defer srv.Close()

	transport := http.DefaultTransport.(*http.Transport).Clone()
	dialer := &net.Dialer{}
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, srv.Listener.Addr().String())
	}
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	provider := newMemoryCacheProvider()
	providerName := registerMemoryCacheProvider(t, provider)
	upsertCh := make(chan db.UpsertImageCacheEntryParams, 1)
	queries := testhelpers.NewQuerierStub()
	queries.CreateCommentInSectionForCommenterResult = 42
	queries.UpsertImageCacheEntryFn = func(_ context.Context, arg db.UpsertImageCacheEntryParams) error {
		select {
		case upsertCh <- arg:
		default:
		}
		return nil
	}
	cfg := config.NewRuntimeConfig()
	cfg.ImageCacheProvider = providerName
	cd := NewCoreData(context.Background(), queries, cfg, WithHTTPClient(client))

	if _, err := cd.CreateCommentInSectionForCommenter("forum", "topic", 1, 1, 9, 1, "[image http://93.184.216.34/logo.png]"); err != nil {
		t.Fatalf("CreateCommentInSectionForCommenter: %v", err)
	}
	select {
	case got := <-upsertCh:
		if got.Status != imageCacheStatusReady {
			t.Fatalf("status = %q, want ready", got.Status)
		}
		if !got.SourceUrl.Valid || got.SourceUrl.String != "http://93.184.216.34/logo.png" {
			t.Fatalf("source url = %#v", got.SourceUrl)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for immediate image cache fetch")
	}
}

func TestProcessPendingRemoteImageCacheEntriesRecordsRetryFailure(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	cfg := config.NewRuntimeConfig()
	cfg.ImageCacheFetchMaxRetries = 3
	cfg.ImageCacheFetchRetryDelay = "1m"
	cd := NewCoreData(context.Background(), queries, cfg)
	queries.ListDuePendingImageCacheEntriesReturns = []*db.ImageCacheEntry{{
		ID:         "abcd1234.png",
		SourceUrl:  sql.NullString{String: "http://127.0.0.1/logo.png", Valid: true},
		SourceKind: imageCacheSourceKindRemote,
		Status:     imageCacheStatusPending,
	}}

	if err := cd.ProcessPendingRemoteImageCacheEntries(context.Background(), 5); err != nil {
		t.Fatalf("ProcessPendingRemoteImageCacheEntries: %v", err)
	}
	if len(queries.ListDuePendingImageCacheEntriesCalls) != 1 {
		t.Fatalf("expected pending list call, got %d", len(queries.ListDuePendingImageCacheEntriesCalls))
	}
	listCall := queries.ListDuePendingImageCacheEntriesCalls[0]
	if listCall.RetryCount != 3 {
		t.Fatalf("retry limit = %d, want 3", listCall.RetryCount)
	}
	if !listCall.NextAttemptAt.Valid {
		t.Fatal("expected due-at filter")
	}
	if len(queries.RecordImageCacheFetchFailureCalls) != 1 {
		t.Fatalf("expected failure record, got %d", len(queries.RecordImageCacheFetchFailureCalls))
	}
	got := queries.RecordImageCacheFetchFailureCalls[0]
	if got.ID != "abcd1234.png" {
		t.Fatalf("id = %q", got.ID)
	}
	if got.RetryCount != 3 || got.RetryCount_2 != 3 {
		t.Fatalf("retry limits = %d/%d, want 3/3", got.RetryCount, got.RetryCount_2)
	}
	if !got.ErrorMessage.Valid || got.ErrorMessage.String == "" {
		t.Fatalf("expected error message, got %#v", got.ErrorMessage)
	}
	if !got.LastAttemptAt.Valid {
		t.Fatal("expected last attempt time")
	}
	if !got.NextAttemptAt.Valid {
		t.Fatal("expected next attempt time")
	}
	if got.NextAttemptAt.Time.Before(got.LastAttemptAt.Time.Add(time.Minute - time.Second)) {
		t.Fatalf("next attempt %s too early after %s", got.NextAttemptAt.Time, got.LastAttemptAt.Time)
	}
}

func TestPrepareImageCacheEntryForServeAllowsMissingMetadataWhenExpiryDisabled(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.GetImageCacheEntryErr = sql.ErrNoRows
	cfg := config.NewRuntimeConfig()
	cfg.ImageCacheMode = "none"
	cd := NewCoreData(context.Background(), queries, cfg)

	ok, err := cd.PrepareImageCacheEntryForServe(context.Background(), "abcd1234.png")
	if err != nil {
		t.Fatalf("PrepareImageCacheEntryForServe: %v", err)
	}
	if !ok {
		t.Fatal("expected cache entry to be servable when expiry is disabled")
	}
	if len(queries.GetImageCacheEntryCalls) != 1 {
		t.Fatalf("expected metadata lookup, got %d", len(queries.GetImageCacheEntryCalls))
	}
}

func TestPrepareImageCacheEntryForServeRejectsExpiredRemoteEntryWithoutSource(t *testing.T) {
	now := time.Now().UTC()
	queries := testhelpers.NewQuerierStub()
	queries.GetImageCacheEntryReturns = &db.ImageCacheEntry{
		ID:         "abcd1234.png",
		SourceKind: imageCacheSourceKindRemote,
		Status:     imageCacheStatusReady,
		CreatedAt:  now.Add(-2 * time.Hour),
		ExpiresAt:  sql.NullTime{Time: now.Add(-time.Hour), Valid: true},
	}
	cfg := config.NewRuntimeConfig()
	cfg.ImageCacheMode = "ttl"
	cfg.ImageCacheTTL = "1h"
	cd := NewCoreData(context.Background(), queries, cfg)

	ok, err := cd.PrepareImageCacheEntryForServe(context.Background(), "abcd1234.png")
	if err != nil {
		t.Fatalf("PrepareImageCacheEntryForServe: %v", err)
	}
	if ok {
		t.Fatal("expected expired remote entry without source URL to be rejected")
	}
}
