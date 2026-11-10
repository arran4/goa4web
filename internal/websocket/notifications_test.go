package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/eventbus"
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

func TestLiveEventEnvelopeDoesNotExposeTaskImplementation(t *testing.T) {
	timestamp := time.Date(2026, time.October, 16, 0, 0, 0, 0, time.UTC)
	got := liveEventFromTaskEvent(eventbus.TaskEvent{
		Path:    "/forum/topic/7/thread/42",
		UserID:  9,
		Time:    timestamp,
		Outcome: eventbus.TaskOutcomeSuccess,
		Data:    map[string]any{"notification": map[string]any{"id": 5}},
	})

	if got.Kind != LiveEventKindNotification {
		t.Fatalf("kind = %q, want %q", got.Kind, LiveEventKindNotification)
	}
	if got.Path != "/forum/topic/7/thread/42" || got.ActorUserID != 9 || !got.Time.Equal(timestamp) {
		t.Fatalf("envelope = %#v", got)
	}
	if _, ok := got.Data["notification"]; !ok {
		t.Fatalf("notification data missing from %#v", got.Data)
	}
}

func TestLiveEventsJSRoute(t *testing.T) {
	reg := routerpkg.NewRegistry()
	mod := NewModule(nil, &config.RuntimeConfig{})
	mod.Register(reg)
	r := mux.NewRouter()
	navReg := nav.NewRegistry()
	routerpkg.RegisterRoutes(r, reg, &config.RuntimeConfig{}, navReg)

	req := httptest.NewRequest("GET", "http://example.com/websocket/live-events.js", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/javascript" {
		t.Fatalf("content-type=%q", ct)
	}
	if !strings.Contains(rec.Body.String(), "'/ws/events'") {
		t.Fatalf("live-events asset does not use the generic endpoint: %q", rec.Body.String())
	}

	match := &mux.RouteMatch{}
	if !r.Match(httptest.NewRequest(http.MethodGet, "http://example.com/ws/events", nil), match) {
		t.Fatal("generic live-events endpoint is not registered")
	}
}

func TestNotificationsHandlerInvalidSession(t *testing.T) {
	core.Store = sessions.NewCookieStore([]byte("test"))
	core.SessionName = "test"

	h := NewNotificationsHandler(nil, &config.RuntimeConfig{})
	req := httptest.NewRequest("GET", "http://example.com/ws/notifications", nil)
	req.AddCookie(&http.Cookie{Name: core.SessionName, Value: "bad"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "invalid session") {
		t.Fatalf("body=%q", rec.Body.String())
	}
}

func TestNotificationsHandlerAuthenticationRequired(t *testing.T) {
	core.Store = sessions.NewCookieStore([]byte("test"))
	core.SessionName = "test"
	sess := sessions.NewSession(core.Store, core.SessionName)
	sess.Values = map[any]any{}

	h := NewNotificationsHandler(nil, &config.RuntimeConfig{})
	req := httptest.NewRequest("GET", "http://example.com/ws/notifications", nil)
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), sess)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "authentication required") {
		t.Fatalf("body=%q", rec.Body.String())
	}
}
