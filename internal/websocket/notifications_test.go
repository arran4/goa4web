package websocket

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	nav "github.com/arran4/goa4web/internal/navigation"
	routerpkg "github.com/arran4/goa4web/internal/router"
)

func TestNotificationsHandlerCheckOriginConfig(t *testing.T) {
	cfg := config.RuntimeConfig{HTTPHostname: "http://example.com"}
	h := NewNotificationsHandler(nil, &cfg)
	req := httptest.NewRequest("GET", "http://example.com/ws/notifications", nil)
	req.Header.Set("Origin", "http://example.com")
	if !h.Upgrader.CheckOrigin(req) {
		t.Fatal("origin from config should be allowed")
	}
}

func TestNotificationsHandlerCheckOriginMultipleHosts(t *testing.T) {
	cfg := config.RuntimeConfig{HTTPHostname: "http://example.com, http://other.com"}
	h := NewNotificationsHandler(nil, &cfg)
	req := httptest.NewRequest("GET", "http://other.com/ws/notifications", nil)
	req.Header.Set("Origin", "http://other.com")
	if !h.Upgrader.CheckOrigin(req) {
		t.Fatal("origin from second config host should be allowed")
	}
}

func TestNotificationsHandlerCheckOriginHostHeader(t *testing.T) {
	cfg := config.RuntimeConfig{HTTPHostname: "http://other.com"}
	h := NewNotificationsHandler(nil, &cfg)
	req := httptest.NewRequest("GET", "http://host/ws/notifications", nil)
	req.Host = "host"
	req.Header.Set("Origin", "http://host")
	if !h.Upgrader.CheckOrigin(req) {
		t.Fatal("origin matching host header should be allowed")
	}
}

func TestNotificationsHandlerCheckOriginDenied(t *testing.T) {
	cfg := config.RuntimeConfig{HTTPHostname: "http://example.com"}
	h := NewNotificationsHandler(nil, &cfg)
	req := httptest.NewRequest("GET", "http://example.com/ws/notifications", nil)
	req.Header.Set("Origin", "http://bad.com")
	if h.Upgrader.CheckOrigin(req) {
		t.Fatal("mismatched origin should be denied")
	}
}

func TestNotificationsJSRoute(t *testing.T) {
	reg := routerpkg.NewRegistry()
	mod := NewModule(nil, &config.RuntimeConfig{})
	mod.Register(reg)
	r := mux.NewRouter()
	navReg := nav.NewRegistry()
	routerpkg.RegisterRoutes(r, reg, &config.RuntimeConfig{}, navReg)

	req := httptest.NewRequest("GET", "http://example.com/websocket/notifications.js", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/javascript" {
		t.Fatalf("content-type=%q", ct)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("empty body")
	}
}
