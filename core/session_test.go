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

const sessionName = "test-session"

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
