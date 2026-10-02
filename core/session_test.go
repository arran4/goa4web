package core_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core"
	"github.com/gorilla/sessions"
)

// sessionName identifies the test CookieStore session.
const sessionName = "test-session"

func TestGetBrowserIDFollowsSessionCookiePolicy(t *testing.T) {
	tests := []struct {
		name     string
		options  *sessions.Options
		secure   bool
		sameSite http.SameSite
		path     string
		httpOnly bool
	}{
		{
			name:     "production",
			options:  &sessions.Options{Path: "/app", HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode},
			secure:   true,
			sameSite: http.SameSiteStrictMode,
			path:     "/app",
			httpOnly: true,
		},
		{
			name:     "local HTTP",
			options:  &sessions.Options{Path: "/", HttpOnly: false, Secure: false, SameSite: http.SameSiteLaxMode},
			secure:   false,
			sameSite: http.SameSiteLaxMode,
			path:     "/",
			httpOnly: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := sessions.NewCookieStore([]byte("test"))
			store.Options = test.options
			core.Store = store
			req := httptest.NewRequest(http.MethodGet, "http://example.com/app", nil)
			recorder := httptest.NewRecorder()
			browserID := core.GetBrowserID(recorder, req)
			if len(browserID) != 64 {
				t.Fatalf("browser ID hash length = %d, want 64", len(browserID))
			}
			cookies := recorder.Result().Cookies()
			if len(cookies) != 1 {
				t.Fatalf("Set-Cookie count = %d, want 1", len(cookies))
			}
			cookie := cookies[0]
			if cookie.Path != test.path || cookie.Secure != test.secure || cookie.HttpOnly != test.httpOnly || cookie.SameSite != test.sameSite {
				t.Fatalf("browser cookie policy = path %q secure %t httpOnly %t sameSite %v", cookie.Path, cookie.Secure, cookie.HttpOnly, cookie.SameSite)
			}
			if cookie.MaxAge <= 0 {
				t.Fatalf("browser cookie MaxAge = %d, want persistent", cookie.MaxAge)
			}
		})
	}
}

func TestGetSessionContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	session := &sessions.Session{Values: map[any]any{"foo": "bar"}}
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), session)
	req = req.WithContext(ctx)

	sess, err := core.GetSession(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess == nil {
		t.Fatal("expected session, got nil")
	}
	if sess.Values["foo"] != "bar" {
		t.Errorf("expected 'bar', got %v", sess.Values["foo"])
	}
}

func TestGetSessionStore(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = sessionName
	req := httptest.NewRequest("GET", "/", nil)
	sess, err := core.GetSession(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess == nil {
		t.Fatal("expected session, got nil")
	}
}

func TestSessionErrorRedirect(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = sessionName
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	core.SessionErrorRedirect(rr, req, nil)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect got %d", rr.Code)
	}
	sc := rr.Header().Get("Set-Cookie")
	if !strings.Contains(sc, "Max-Age=0") {
		t.Errorf("expected cleared cookie, got %q", sc)
	}
	loc := rr.Header().Get("Location")
	if loc != "/login?back=%2F" {
		t.Errorf("unexpected location %q", loc)
	}
}

func TestGetSessionOrFailBadSession(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = sessionName
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionName, Value: "bad"})
	rr := httptest.NewRecorder()
	sess, ok := core.GetSessionOrFail(rr, req)
	if ok {
		t.Fatalf("expected failure, got session %v", sess)
	}
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect got %d", rr.Code)
	}
	sc := rr.Header().Get("Set-Cookie")
	if !strings.Contains(sc, "Max-Age=0") {
		t.Errorf("expected cleared cookie, got %q", sc)
	}
	loc := rr.Header().Get("Location")
	if loc != "/login?back=%2F" {
		t.Errorf("unexpected location %q", loc)
	}
}

func TestGetSessionOrFail(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = sessionName
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	sess, ok := core.GetSessionOrFail(rr, req)
	if !ok {
		t.Fatalf("expected success")
	}
	if sess == nil {
		t.Fatal("expected session")
	}
}
