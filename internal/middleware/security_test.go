package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func newCoreData(t *testing.T, cfg config.RuntimeConfig) (*common.CoreData, *db.QuerierStub) {
	t.Helper()
	stub := testhelpers.NewQuerierStub()
	stub.ListActiveBansReturns = []*db.BannedIp{}
	if cfg.HSTSHeaderValue == "" {
		cfg.HSTSHeaderValue = "max-age=63072000; includeSubDomains"
	}
	cd := common.NewCoreData(context.Background(), stub, &cfg)
	return cd, stub
}

func TestSecurityHeadersMiddlewareHTTP(t *testing.T) {
	cfg := config.RuntimeConfig{HTTPHostname: "http://example.com"}
	cd, stub := newCoreData(t, cfg)

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler := SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	handler.ServeHTTP(rec, req)

	if h := rec.Header().Get("Strict-Transport-Security"); h != "" {
		t.Fatalf("unexpected HSTS header %q", h)
	}

	if stub.ListActiveBansCalls != 1 {
		t.Fatalf("ListActiveBansCalls=%d want 1", stub.ListActiveBansCalls)
	}
}

func TestSecurityHeadersMiddlewareHTTPS(t *testing.T) {
	cfg := config.RuntimeConfig{HTTPHostname: "https://example.com", BaseURL: "https://example.com"}
	cd, stub := newCoreData(t, cfg)

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler := SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	handler.ServeHTTP(rec, req)

	wantHSTS := cd.Config.HSTSHeaderValue
	if h := rec.Header().Get("Strict-Transport-Security"); h != wantHSTS {
		t.Fatalf("unexpected HSTS header %q", h)
	}

	if stub.ListActiveBansCalls != 1 {
		t.Fatalf("ListActiveBansCalls=%d want 1", stub.ListActiveBansCalls)
	}
}

func TestSecurityHeadersMiddlewareForwardedProto(t *testing.T) {
	cfg := config.RuntimeConfig{HTTPHostname: "http://example.com"}
	cd, stub := newCoreData(t, cfg)

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler := SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	handler.ServeHTTP(rec, req)

	wantHSTS := cd.Config.HSTSHeaderValue
	if h := rec.Header().Get("Strict-Transport-Security"); h != wantHSTS {
		t.Fatalf("unexpected HSTS header %q", h)
	}

	if stub.ListActiveBansCalls != 1 {
		t.Fatalf("ListActiveBansCalls=%d want 1", stub.ListActiveBansCalls)
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	handler := SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("X-Frame-Options=%q want DENY", got)
	}
	if got := rr.Header().Get("Referrer-Policy"); got != "no-referrer" {
		t.Fatalf("Referrer-Policy=%q want no-referrer", got)
	}
}
