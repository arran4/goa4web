package user

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	gorillaws "github.com/gorilla/websocket"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/websocket"
)

type fragmentTestQuerier struct {
	*db.QuerierStub
	notifications []*db.Notification
	unreadCount   int64
}

func (q *fragmentTestQuerier) GetUnreadNotificationCountForLister(_ context.Context, _ int32) (int64, error) {
	return q.unreadCount, nil
}

func (q *fragmentTestQuerier) ListUnreadNotificationsForLister(_ context.Context, _ db.ListUnreadNotificationsForListerParams) ([]*db.Notification, error) {
	return q.notifications, nil
}

func (q *fragmentTestQuerier) ListNotificationsForLister(_ context.Context, _ db.ListNotificationsForListerParams) ([]*db.Notification, error) {
	return q.notifications, nil
}

func (q *fragmentTestQuerier) GetNotificationCountForLister(_ context.Context, _ int32) (int64, error) {
	return int64(len(q.notifications)), nil
}

func (q *fragmentTestQuerier) ListSubscriptionsByUser(_ context.Context, _ int32) ([]*db.ListSubscriptionsByUserRow, error) {
	return []*db.ListSubscriptionsByUserRow{
		{Method: "internal", Pattern: "reply:/forum/topic/*"},
	}, nil
}

func withAuthenticatedUserContext(r *http.Request, uid int32, q db.Querier, cfg *config.RuntimeConfig) *http.Request {
	store := sessions.NewCookieStore([]byte("test-secret-key-for-sessions-32b"))
	core.Store = store
	core.SessionName = "goa4web-session"
	sess := sessions.NewSession(store, core.SessionName)
	sess.Values["UID"] = uid
	ctx := context.WithValue(r.Context(), core.ContextValues("session"), sess)
	cd := common.NewCoreData(ctx, q, cfg)
	cd.UserID = uid
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	return r.WithContext(ctx)
}

func TestNotificationFragmentsAndXSSSafety(t *testing.T) {
	cfg := &config.RuntimeConfig{
		NotificationsEnabled: true,
		HTTPHostname:         "http://example.com",
	}
	q := &fragmentTestQuerier{
		QuerierStub: &db.QuerierStub{},
		unreadCount: 1,
		notifications: []*db.Notification{
			{
				ID:        77,
				CreatedAt: time.Date(2026, time.November, 15, 10, 0, 0, 0, time.UTC),
				Message:   sql.NullString{String: `<script>alert("xss")</script><img src=x onerror=alert(1)>`, Valid: true},
				Link:      sql.NullString{String: `javascript:alert(1)" onclick="alert(2)`, Valid: true},
			},
		},
	}

	r := mux.NewRouter()
	RegisterRoutes(r, cfg)

	t.Run("unauthenticated fragment requests are rejected", func(t *testing.T) {
		for _, path := range []string{"/usr/notifications/fragment/list", "/usr/notifications/fragment/count"} {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			cd := common.NewCoreData(req.Context(), q, cfg)
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code == http.StatusOK && strings.Contains(rec.Body.String(), "notif-77") {
				t.Fatalf("%s leaked authenticated fragment to anonymous caller", path)
			}
		}
	})

	t.Run("authenticated list fragment escapes malicious message and link", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/usr/notifications/fragment/list", nil)
		req = withAuthenticatedUserContext(req, 42, q, cfg)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		body := rec.Body.String()
		if strings.Contains(body, "<script>") || strings.Contains(body, "<img") || strings.Contains(body, "javascript:alert") {
			t.Fatalf("unescaped XSS payload rendered in fragment: %s", body)
		}
		if !strings.Contains(body, "&lt;script&gt;") {
			t.Fatalf("expected HTML-escaped script tag in fragment: %s", body)
		}
		if !strings.Contains(body, `href="/usr/notifications/open/77"`) {
			t.Fatalf("expected safe /usr/notifications/open/77 link in fragment: %s", body)
		}
	})

	t.Run("authenticated count fragment includes declarative live attributes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/usr/notifications/fragment/count", nil)
		req = withAuthenticatedUserContext(req, 42, q, cfg)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		body := rec.Body.String()
		for _, want := range []string{
			`id="notif-index"`,
			`data-live-event="notification"`,
			`data-live-url="/usr/notifications/fragment/count"`,
			`data-live-swap="outerHTML"`,
			`(1)`,
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("count fragment missing %q: %s", want, body)
			}
		}
	})

	t.Run("websocket event triggers authorized fragment convergence", func(t *testing.T) {
		bus := eventbus.NewBus()
		wsHandler := websocket.NewNotificationsHandler(bus, cfg)

		store := sessions.NewCookieStore([]byte("test-secret-key-for-sessions-32b"))
		core.Store = store
		core.SessionName = "goa4web-session"

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			sess := sessions.NewSession(store, core.SessionName)
			sess.Values["UID"] = int32(42)
			ctx := context.WithValue(req.Context(), core.ContextValues("session"), sess)
			cd := common.NewCoreData(ctx, q, cfg)
			cd.UserID = 42
			ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
			req = req.WithContext(ctx)
			if req.URL.Path == "/ws/events" {
				wsHandler.ServeHTTP(w, req)
				return
			}
			r.ServeHTTP(w, req)
		}))
		defer srv.Close()

		wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/events"
		conn, _, err := gorillaws.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("dial ws: %v", err)
		}
		defer func() { _ = conn.Close() }()

		marker := time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC)
		if err := bus.Publish(eventbus.NotificationsChanged{RecipientID: 99, Time: marker.Add(-time.Minute)}); err != nil {
			t.Fatalf("publish unrelated notification signal: %v", err)
		}
		if err := bus.Publish(eventbus.TaskEvent{
			Path:   "/forum/topic/1",
			UserID: 99,
			Task:   replyTaskStub{},
			Data: map[string]any{
				"notification": map[string]any{"id": 77},
			},
		}); err != nil {
			t.Fatalf("publish task event: %v", err)
		}
		if err := bus.Publish(eventbus.NotificationsChanged{RecipientID: 42, Time: marker}); err != nil {
			t.Fatalf("publish own notification signal: %v", err)
		}

		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		var evt websocket.LiveEvent
		if err := conn.ReadJSON(&evt); err != nil {
			t.Fatalf("read ws event: %v", err)
		}
		if evt.Kind != websocket.LiveEventKindNotification || !evt.Time.Equal(marker) {
			t.Fatalf("received unexpected event: %#v", evt)
		}

		resp, err := srv.Client().Get(srv.URL + "/usr/notifications/fragment/list")
		if err != nil {
			t.Fatalf("GET fragment: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("fragment status = %d", resp.StatusCode)
		}
	})
}

type replyTaskStub struct{}

func (replyTaskStub) Name() string                                  { return "Reply" }
func (replyTaskStub) Action(http.ResponseWriter, *http.Request) any { return nil }
