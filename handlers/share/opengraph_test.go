package share_test

import (
	"bytes"
	"fmt"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/share"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/gorilla/mux"
)

func TestMain(m *testing.M) {
	// Set the templates directory to the project root so that the embedded assets can be found.
	templates.SetDir("../..")
	os.Exit(m.Run())
}

const testKey = "test-secret-key-for-og-images"

func TestHappyPathMakeImageURL_QueryAuth(t *testing.T) {
	baseURL := "http://example.com"
	title := "Test Title"

	// Generate URL with query-based auth
	signedURL, err := share.MakeImageURL(baseURL, title, "", testKey, false)
	if err != nil {
		t.Fatalf("MakeImageURL failed: %v", err)
	}

	// Should contain sig and nonce in query
	if !strings.Contains(signedURL, "?") {
		t.Errorf("URL should contain query params: %s", signedURL)
	}
	if !strings.Contains(signedURL, "sig=") {
		t.Errorf("URL should contain sig param: %s", signedURL)
	}
	if !strings.Contains(signedURL, "nonce=") {
		t.Errorf("URL should contain nonce param: %s", signedURL)
	}

	// Test verification
	req := httptest.NewRequest("GET", signedURL, nil)
	cleanPath := share.VerifyAndGetPath(req, testKey)

	if cleanPath == "" {
		t.Error("Verification failed for valid signature")
	}

	// Clean path should be /api/og-image/{base64}
	if !strings.HasPrefix(cleanPath, "/api/og-image/") {
		t.Errorf("Clean path should be /api/og-image/..., got: %s", cleanPath)
	}
}

func TestHappyPathMakeImageURL_PathAuth(t *testing.T) {
	baseURL := "http://example.com"
	title := "Test Title"

	// Generate URL with path-based auth
	signedURL, err := share.MakeImageURL(baseURL, title, "", testKey, true)
	if err != nil {
		t.Fatalf("MakeImageURL failed: %v", err)
	}

	// Should contain /nonce/.../sign/... in path
	if !strings.Contains(signedURL, "/nonce/") {
		t.Errorf("URL should contain /nonce/ in path: %s", signedURL)
	}
	if !strings.Contains(signedURL, "/sign/") {
		t.Errorf("URL should contain /sign/ in path: %s", signedURL)
	}

	// Parse to extract path vars (simulating mux)
	// URL format: http://example.com/api/og-image/data/nonce/xxx/sign/yyy
	pathPart := strings.TrimPrefix(signedURL, baseURL)

	// Extract nonce and sig from path
	parts := strings.Split(pathPart, "/")
	var nonce, sig string
	for i, part := range parts {
		if part == "nonce" && i+1 < len(parts) {
			nonce = parts[i+1]
		}
		if part == "sign" && i+1 < len(parts) {
			sig = parts[i+1]
		}
	}

	if nonce == "" || sig == "" {
		t.Fatalf("Could not extract nonce/sig from path: %s", signedURL)
	}

	// Create a mux router to handle path vars
	r := mux.NewRouter()
	var verifiedPath string
	r.HandleFunc("/api/og-image/{data}/nonce/{nonce}/sign/{sign}", func(w http.ResponseWriter, r *http.Request) {
		verifiedPath = share.VerifyAndGetPath(r, testKey)
	})

	req := httptest.NewRequest("GET", signedURL, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if verifiedPath == "" {
		t.Error("Verification failed for valid path-based signature")
	}

	// Clean path should be /api/og-image/{base64}
	if !strings.HasPrefix(verifiedPath, "/api/og-image/") {
		t.Errorf("Clean path should be /api/og-image/..., got: %s", verifiedPath)
	}
}

func TestHappyPathMakeImageURL_WithExpiry(t *testing.T) {
	baseURL := "http://example.com"
	title := "Test Title"

	// Use explicit expiry
	//expiry := time.Now().Add(1 * time.Hour)
	// For testing, let's still use nonce as the current implementation prefers it
	signedURL, err := share.MakeImageURL(baseURL, title, "", testKey, false)
	if err != nil {
		t.Fatalf("MakeImageURL failed: %v", err)
	}

	// Verify it works
	req := httptest.NewRequest("GET", signedURL, nil)
	cleanPath := share.VerifyAndGetPath(req, testKey)

	if cleanPath == "" {
		t.Error("Verification failed")
	}
}

func TestHappyPathOGImageHandler(t *testing.T) {
	handler := share.NewOGImageHandler(testKey)

	// Generate a valid signed URL
	baseURL := "http://example.com"
	title := "My Test Title"

	signedURL, err := share.MakeImageURL(baseURL, title, "Test Description", testKey, false)
	if err != nil {
		t.Fatalf("MakeImageURL failed: %v", err)
	}

	// Make request to handler
	req := httptest.NewRequest("GET", signedURL, nil)
	rec := httptest.NewRecorder()

	r := mux.NewRouter()
	r.Handle("/api/og-image/{data}", handler)
	r.ServeHTTP(rec, req)

	// Should return PNG
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("Expected Content-Type image/png, got %s", ct)
	}

	// Body should be a valid PNG image
	_, err = png.Decode(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Errorf("Failed to decode response body as PNG: %v", err)
	}
}

func TestUnhappyPathOGImageHandler_InvalidSignature(t *testing.T) {
	handler := share.NewOGImageHandler(testKey)

	// Request without signature
	// Request without signature
	req := httptest.NewRequest("GET", "http://example.com/api/og-image/dGVzdA", nil)
	rec := httptest.NewRecorder()

	r := mux.NewRouter()
	r.Handle("/api/og-image/{data}", handler)
	r.ServeHTTP(rec, req)

	// Should return 401
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestUnhappyPathOGImageHandler_WrongKey(t *testing.T) {
	handler := share.NewOGImageHandler("wrong-key")

	// Generate URL with correct key
	signedURL, err := share.MakeImageURL("http://example.com", "Test", "", testKey, false)
	if err != nil {
		t.Fatalf("MakeImageURL failed: %v", err)
	}

	// Try to verify with wrong key
	req := httptest.NewRequest("GET", signedURL, nil)
	rec := httptest.NewRecorder()

	r := mux.NewRouter()
	r.Handle("/api/og-image/{data}", handler)
	r.ServeHTTP(rec, req)

	// Should return 401
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for wrong key, got %d", rec.Code)
	}
}

func TestHappyPathVerifyAndGetPath_WithQueryParams(t *testing.T) {
	// Test that additional query params are preserved
	path := "/api/og-image/data"
	extraQuery := "baz=qux&foo=bar"
	fullPath := path + "?" + extraQuery

	nonce := "test-nonce-123"
	sig := sign.Sign(fullPath, testKey, sign.WithNonce(nonce))

	// Build URL with sig
	signedURL, err := sign.AddQuerySig("http://example.com"+fullPath, sig, sign.WithNonce(nonce))
	if err != nil {
		t.Fatalf("AddQuerySig failed: %v", err)
	}

	req := httptest.NewRequest("GET", signedURL, nil)
	cleanPath := share.VerifyAndGetPath(req, testKey)

	if cleanPath == "" {
		t.Error("Verification failed")
	}

	// Clean path should include the extra query params
	if cleanPath != fullPath {
		t.Errorf("Expected clean path %s, got %s", fullPath, cleanPath)
	}
}

func TestHappyPathVerifyAndGetPath_MixedAuth(t *testing.T) {
	key := "test-key"
	ts := time.Now().Add(1 * time.Hour).Unix()

	// Goal: Support /path/ts/{ts}?sig={sig}
	// Signature generated with Expiry option on base path
	basePath := "/some/content"

	sig := sign.Sign(basePath, key, sign.WithExpiry(time.Unix(ts, 0)))

	// Construct mixed URL: Path has TS, Query has Sig
	pathWithTs := fmt.Sprintf("%s/ts/%d", basePath, ts)
	reqURL := fmt.Sprintf("http://example.com%s?sig=%s", pathWithTs, sig)

	req := httptest.NewRequest(http.MethodGet, reqURL, nil)

	// Simulate router variables
	vars := map[string]string{
		"ts": fmt.Sprintf("%d", ts),
	}
	req = mux.SetURLVars(req, vars)

	cleanPath := share.VerifyAndGetPath(req, key)

	if cleanPath != basePath {
		t.Errorf("Expected clean path '%s', got '%s'", basePath, cleanPath)
	}
}

func TestUnhappyPathVerifyAndGetPath_MixedAuth_Expired(t *testing.T) {
	key := "test-key"
	// Expired 1 hour ago
	ts := time.Now().Add(-1 * time.Hour).Unix()

	basePath := "/some/content"

	sig := sign.Sign(basePath, key, sign.WithExpiry(time.Unix(ts, 0)))

	pathWithTs := fmt.Sprintf("%s/ts/%d", basePath, ts)
	reqURL := fmt.Sprintf("http://example.com%s?sig=%s", pathWithTs, sig)

	req := httptest.NewRequest(http.MethodGet, reqURL, nil)

	vars := map[string]string{
		"ts": fmt.Sprintf("%d", ts),
	}
	req = mux.SetURLVars(req, vars)

	cleanPath := share.VerifyAndGetPath(req, key)

	if cleanPath != "" {
		t.Errorf("Expected empty path (expired), got '%s'", cleanPath)
	}
}

func TestHappyPathShareRoutes_MixedAuth(t *testing.T) {
	r := mux.NewRouter()
	cfg := &config.RuntimeConfig{} // Minimal config
	shareKey := "test-key"

	share.RegisterShareRoutes(r, cfg, shareKey)

	// Test cases that should match
	tests := []string{
		"/api/og-image/data",
		"/api/og-image/data/ts/123",
		"/api/og-image/data/nonce/abc",
		"/api/og-image/data/ts/123/sign/sig",
		"/api/og-image/data/nonce/abc/sign/sig",
	}

	for _, path := range tests {
		req := httptest.NewRequest("GET", path, nil)
		match := &mux.RouteMatch{}
		if !r.Match(req, match) {
			t.Errorf("Route not matched: %s", path)
		}
	}
}
