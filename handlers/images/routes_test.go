package images

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	intimages "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/sign"
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
	// Create a dummy image directory or mock the serving part?
	// For this test, we only care about middleware passing (200 OK or 404 NotFound if file missing, but not 403 Forbidden).
	// But serveImage will try to serve file.
	// We can't easily mock serveImage without changing code, but if middleware passes, serveImage runs.
	// serveImage checks ValidID which we will use a valid ID.
	// It then tries to serve file. If file missing -> 404.
	// If middleware fails -> 403.
	// So we expect !403.

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
	// If the middleware works, it should reach serveImage, which likely returns 404 for missing file.
	// If middleware fails, it returns 403.
	if rr.Code == http.StatusForbidden {
		t.Errorf("Request was forbidden (403). Middleware failed verification. URL: %s", signedURLStr)
	} else if rr.Code != http.StatusNotFound && rr.Code != http.StatusOK {
		t.Errorf("Unexpected status code: %d. Expected 404 (file missing) or 200 (if we mocked file).", rr.Code)
	} else {
		t.Logf("Success: Got status %d (likely passed middleware)", rr.Code)
	}
}
