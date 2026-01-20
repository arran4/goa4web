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
