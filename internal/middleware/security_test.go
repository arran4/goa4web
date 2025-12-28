package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func newCoreData(t *testing.T, cfg config.RuntimeConfig) (*common.CoreData, sqlmock.Sqlmock, func()) {
	t.Helper()
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	cleanup := func() { conn.Close() }
	queries := db.New(conn)
	if cfg.HSTSHeaderValue == "" {
		cfg.HSTSHeaderValue = "max-age=63072000; includeSubDomains"
	}
	cd := common.NewCoreData(context.Background(), queries, &cfg)
	return cd, mock, cleanup
}

func TestSecurityHeadersMiddlewareHTTP(t *testing.T) {
	cfg := config.RuntimeConfig{HTTPHostname: "http://example.com"}
	cd, mock, cleanup := newCoreData(t, cfg)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("FROM banned_ips")).WillReturnRows(sqlmock.NewRows([]string{"id", "ip_net", "reason", "created_at", "expires_at", "canceled_at"}))

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler := SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	handler.ServeHTTP(rec, req)

	if h := rec.Header().Get("Strict-Transport-Security"); h != "" {
		t.Fatalf("unexpected HSTS header %q", h)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSecurityHeadersMiddlewareHTTPS(t *testing.T) {
	cfg := config.RuntimeConfig{HTTPHostname: "https://example.com"}
	cd, mock, cleanup := newCoreData(t, cfg)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("FROM banned_ips")).WillReturnRows(sqlmock.NewRows([]string{"id", "ip_net", "reason", "created_at", "expires_at", "canceled_at"}))

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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSecurityHeadersMiddlewareForwardedProto(t *testing.T) {
	cfg := config.RuntimeConfig{HTTPHostname: "http://example.com"}
	cd, mock, cleanup := newCoreData(t, cfg)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("FROM banned_ips")).WillReturnRows(sqlmock.NewRows([]string{"id", "ip_net", "reason", "created_at", "expires_at", "canceled_at"}))

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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
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
