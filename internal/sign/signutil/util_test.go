package signutil_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/sign/signutil"
	"github.com/gorilla/mux"
)

func TestGetSignedData_MixedAuth(t *testing.T) {
	key := "test-key"
	ts := time.Now().Add(1 * time.Hour).Unix()

	basePath := "/api/og-image/data"

	sig := sign.Sign(basePath, key, sign.WithExpiry(time.Unix(ts, 0)))

	// Mixed: Path has TS, Query has Sig
	pathWithTs := fmt.Sprintf("%s/ts/%d", basePath, ts)
	reqURL := fmt.Sprintf("http://example.com%s?sig=%s", pathWithTs, sig)

	req := httptest.NewRequest(http.MethodGet, reqURL, nil)

	vars := map[string]string{
		"ts": fmt.Sprintf("%d", ts),
	}
	req = mux.SetURLVars(req, vars)

	signedData, err := signutil.GetSignedData(req, key)
	if err != nil {
		t.Fatalf("GetSignedData returned error: %v", err)
	}
	if !signedData.Valid {
		t.Error("Expected Valid=true, got false")
	}
}

func TestGetSignedData_MixedAuth_Expired(t *testing.T) {
	key := "test-key"
	ts := time.Now().Add(-1 * time.Hour).Unix()

	basePath := "/api/og-image/data"

	sig := sign.Sign(basePath, key, sign.WithExpiry(time.Unix(ts, 0)))

	pathWithTs := fmt.Sprintf("%s/ts/%d", basePath, ts)
	reqURL := fmt.Sprintf("http://example.com%s?sig=%s", pathWithTs, sig)

	req := httptest.NewRequest(http.MethodGet, reqURL, nil)

	vars := map[string]string{
		"ts": fmt.Sprintf("%d", ts),
	}
	req = mux.SetURLVars(req, vars)

	signedData, err := signutil.GetSignedData(req, key)
	if err != nil {
		t.Fatalf("GetSignedData returned error: %v", err)
	}
	if signedData.Valid {
		t.Error("Expected Valid=false (expired), got true")
	}
}

func TestRemoveShared(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"Basic path", "/private/shared/topic/1", "/private/topic/1"},
		{"Nested path", "/forum/shared/topic/1/thread/2", "/forum/topic/1/thread/2"},
		{"Query params untouched", "/private/shared/topic/1?foo=bar&shared=true", "/private/topic/1?foo=bar&shared=true"},
		{"Percent encoding untouched", "/private/shared/topic/1?foo=%20bar", "/private/topic/1?foo=%20bar"},
		{"Without shared", "/private/topic/1", "/private/topic/1"},
		{"Empty path", "", ""},
		{"Root path", "/", "/"},
		{"Short path", "/private", "/private"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := signutil.RemoveShared(tt.path); got != tt.expected {
				t.Errorf("RemoveShared() = %v, want %v", got, tt.expected)
			}
		})
	}
}
