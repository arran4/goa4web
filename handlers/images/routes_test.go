package images

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	intimages "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/upload"
	"github.com/arran4/goa4web/internal/upload/local"
)

func TestValidID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"abcd", true},
		{"1234", true},
		{"a1b2c3", true},
		{"a.b", false},
		{"abc!", false},
		{".", false},
		{"..", false},
		{"hi/hi", false},
		{"text.text", true},
		{"a/bc", false},
		{"abc", false},
	}
	for _, tt := range tests {
		if got := intimages.ValidID(tt.id); got != tt.valid {
			t.Errorf("ValidID(%q) = %v want %v", tt.id, got, tt.valid)
		}
	}
}

func TestImageRouteInvalidID(t *testing.T) {
	r := mux.NewRouter()
	cfg := config.NewRuntimeConfig()
	navReg := navigation.NewRegistry()
	RegisterRoutes(r, cfg, navReg)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/images/image/abc!", nil)
	cd := common.NewCoreData(req.Context(), nil, cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)

	r.ServeHTTP(rr, req.WithContext(ctx))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("want %d got %d", http.StatusForbidden, rr.Code)
	}
}

func TestCacheRouteInvalidID(t *testing.T) {
	r := mux.NewRouter()
	cfg := config.NewRuntimeConfig()
	navReg := navigation.NewRegistry()
	RegisterRoutes(r, cfg, navReg)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/images/cache/abc!", nil)
	cd := common.NewCoreData(req.Context(), nil, cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)

	r.ServeHTTP(rr, req.WithContext(ctx))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("want %d got %d", http.StatusForbidden, rr.Code)
	}
}

func TestVerifyMiddlewareUnauthorized(t *testing.T) {
	called := false
	h := verifyMiddleware("image:")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	req := httptest.NewRequest("GET", "/images/image/abcd", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abcd"})
	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req.WithContext(ctx))

	if called {
		t.Fatalf("next handler was called")
	}
	if rr.Code != http.StatusForbidden {
		t.Fatalf("want %d got %d", http.StatusForbidden, rr.Code)
	}
}

func TestVerifyMiddlewareAllowsQuerySignedImage(t *testing.T) {
	called := false
	cfg := config.NewRuntimeConfig()
	cfg.BaseURL = "http://localhost"
	key := "k"
	signedURL := "http://localhost/images/image/abcd.png?size=small&sig=" + sign.Sign("image:abcd.png?size=small", key, sign.WithOutNonce())
	parsed, err := url.Parse(signedURL)
	if err != nil {
		t.Fatalf("parse signed url: %v", err)
	}
	h := verifyMiddleware("image:")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	req := httptest.NewRequest("GET", parsed.Path+"?"+parsed.RawQuery, nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abcd.png"})
	cd := common.NewCoreData(req.Context(), nil, cfg, common.WithImageSignKey(key))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req.WithContext(ctx))

	if !called {
		t.Fatalf("next handler was not called")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("want %d got %d", http.StatusOK, rr.Code)
	}
}

func TestSignImageURL_EndToEnd(t *testing.T) {
	// Setup Router
	r := mux.NewRouter()
	cfg := config.NewRuntimeConfig()
	cfg.BaseURL = "http://localhost"
	navReg := navigation.NewRegistry()
	RegisterRoutes(r, cfg, navReg)

	// Setup CoreData with key
	req := httptest.NewRequest("GET", "/", nil)
	key := "test-image-key"
	cd := common.NewCoreData(req.Context(), nil, cfg, common.WithImageSignKey(key))

	// Generate Signed URL
	// We use a valid ID format
	imageID := "ccf0e454-e774-4d6f-9ac3-2d6c15baad6d.png" // Valid format from user report
	signedURLStr := cd.SignImageURL("image:"+imageID, 1*time.Hour)

	t.Logf("Signed URL: %s", signedURLStr)

	u, err := url.Parse(signedURLStr)
	if err != nil {
		t.Fatalf("Failed to parse generated URL: %v", err)
	}

	// Create Request
	req = httptest.NewRequest("GET", u.Path+"?"+u.RawQuery, nil)
	// Must inject CoreData into context as middleware expects it
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// We expect 404 (because file doesn't exist) but NOT 403 (Forbidden).
	if rr.Code == http.StatusForbidden {
		t.Errorf("Request was forbidden (403). Middleware failed verification. URL: %s", signedURLStr)
	} else if rr.Code != http.StatusNotFound && rr.Code != http.StatusOK {
		t.Errorf("Unexpected status code: %d. Expected 404 (file missing) or 200 (if we mocked file).", rr.Code)
	} else {
		t.Logf("Success: Got status %d (likely passed middleware)", rr.Code)
	}
}

func TestThumbnailRegeneration(t *testing.T) {
	// 1. Setup
	local.Register()

	tmpDir, err := os.MkdirTemp("", "images-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	uploadDir := filepath.Join(tmpDir, "uploads")
	cacheDir := filepath.Join(tmpDir, "cache")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.NewRuntimeConfig()
	cfg.ImageUploadProvider = "local"
	cfg.ImageCacheProvider = "local"
	cfg.ImageUploadDir = uploadDir
	cfg.ImageCacheDir = cacheDir
	cfg.BaseURL = "http://localhost"

	req := httptest.NewRequest("GET", "/", nil)
	key := "test-key"
	cd := common.NewCoreData(req.Context(), nil, cfg, common.WithImageSignKey(key))

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	imgData := buf.Bytes()

	// Save original image using provider
	id := "test1234"
	ext := ".png"
	sub1, sub2 := id[:2], id[2:4]
	fname := id + ext

	prov := upload.ProviderFromConfig(cfg)
	if prov == nil {
		t.Fatal("ProviderFromConfig returned nil")
	}

	// Write original
	if err := prov.Write(context.Background(), path.Join(sub1, sub2, fname), imgData); err != nil {
		t.Fatalf("Failed to write original image: %v", err)
	}

	// Ensure cache is EMPTY for this file
	thumbID := id + "_thumb" + ext
	cacheKey := path.Join(sub1, sub2, thumbID)
	// We can try to read from cache provider to ensure it's not there
	cacheProv := upload.CacheProviderFromConfig(cfg)
	if _, err := cacheProv.Read(context.Background(), cacheKey); err == nil {
		t.Fatal("Thumbnail should not exist yet")
	}

	// 2. Setup Router
	r := mux.NewRouter()
	navReg := navigation.NewRegistry()
	RegisterRoutes(r, cfg, navReg)

	// 3. Request Thumbnail
	// Construct signed URL
	signedURLStr := cd.SignCacheURL(thumbID, 1*time.Hour)
	u, err := url.Parse(signedURLStr)
	if err != nil {
		t.Fatalf("Failed to parse signed URL: %v", err)
	}

	req = httptest.NewRequest("GET", u.Path+"?"+u.RawQuery, nil)
	// Must inject CoreData into context
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// 4. Assertions
	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rr.Code)
	}

	// Verify thumbnail was created
	if _, err := cacheProv.Read(context.Background(), cacheKey); err != nil {
		t.Errorf("Thumbnail was not created in cache: %v", err)
	}
}
