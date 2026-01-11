package share

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/sharesign"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/gorilla/mux"
)

func TestMakeImageURLAndVerify(t *testing.T) {
	key := "secret"
	signer := sharesign.NewSigner(&config.RuntimeConfig{HTTPHostname: "http://example.com"}, key)
	baseURL := "http://example.com"
	title := "Test Title"
	encodedTitle := base64.RawURLEncoding.EncodeToString([]byte(title))
	expectedBase := "/api/og-image/" + encodedTitle

	tests := []struct {
		name        string
		usePathAuth bool
	}{
		{
			name:        "Query_Default_Nonce",
			usePathAuth: false,
		},
		{
			name:        "PathAuth_Default_Nonce",
			usePathAuth: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlStr := MakeImageURL(baseURL, title, signer, tt.usePathAuth)

			// Verify URL structure - should always contain nonce
			if !strings.Contains(urlStr, "nonce") {
				t.Errorf("URL expected to contain nonce but didn't: %s", urlStr)
			}

			// Simulate Request
			req, err := http.NewRequest("GET", urlStr, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// If PathAuth, we need to extract vars as Mux would
			if tt.usePathAuth {
				r := mux.NewRouter()
				r.HandleFunc("/api/og-image/{data}/nonce/{nonce}/sign/{sign}", func(w http.ResponseWriter, r *http.Request) {
					path := VerifyAndGetPath(r, signer)
					if path != expectedBase {
						t.Errorf("VerifyAndGetPath returned %s, want %s", path, expectedBase)
					}
				})

				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				if w.Code == 404 {
					t.Errorf("Test route did not match generated URL: %s", urlStr)
				}
			} else {
				// Query param case
				path := VerifyAndGetPath(req, signer)
				if path != expectedBase {
					t.Errorf("VerifyAndGetPath returned %s, want %s", path, expectedBase)
				}
			}
		})
	}
}

func TestVerifyAndGetPath_Failure(t *testing.T) {
	key := "secret"
	signer := sharesign.NewSigner(&config.RuntimeConfig{HTTPHostname: "http://example.com"}, key)

	// Test invalid signature
	urlStr := "http://example.com/api/og-image/data?ts=12345&sig=invalid"
	req, _ := http.NewRequest("GET", urlStr, nil)

	path := VerifyAndGetPath(req, signer)
	if path != "" {
		t.Errorf("Expected empty path for invalid signature, got %s", path)
	}
}

func TestVerifyAndGetPath_WithQueryParams(t *testing.T) {
	key := "secret"
	signer := sharesign.NewSigner(&config.RuntimeConfig{HTTPHostname: "http://example.com"}, key)

	// Case 1: Query Auth with extra params
	// "/test?foo=bar"
	sig := signer.Sign("/test?foo=bar", sign.WithExpiry(time.Now().Add(time.Hour)))
	urlStr := fmt.Sprintf("http://example.com/test?foo=bar&sig=%s", sig)

	req, _ := http.NewRequest("GET", urlStr, nil)
	path := VerifyAndGetPath(req, signer)

	// VerifyAndGetPath reconstructs sorted query. foo=bar.
	expected := "/test?foo=bar"
	if path != expected {
		t.Errorf("QueryAuth: Expected %s, got %s", expected, path)
	}

	// Case 2: Path Auth with extra params
	// "/test?foo=bar"
	sig = signer.Sign("/test?foo=bar", sign.WithExpiry(time.Now().Add(time.Hour)))
	// Route simulation: /test/sign/{sig}?foo=bar
	// We need mux vars.

	r := mux.NewRouter()
	r.HandleFunc("/test/sign/{sign}", func(w http.ResponseWriter, r *http.Request) {
		path := VerifyAndGetPath(r, signer)
		if path != expected {
			t.Errorf("PathAuth: Expected %s, got %s", expected, path)
		}
	})

	urlStrPath := fmt.Sprintf("http://example.com/test/sign/%s?foo=bar", sig)
	reqPath, _ := http.NewRequest("GET", urlStrPath, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqPath)

	if w.Code == 404 {
		t.Errorf("PathAuth: Route not matched")
	}
}
