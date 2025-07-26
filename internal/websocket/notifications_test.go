package websocket

import (
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
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
