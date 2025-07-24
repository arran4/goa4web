package core_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core"
)

var (
	store       *sessions.CookieStore
	sessionName = "test-session"
)

func TestSessionMiddlewareBadSession(t *testing.T) {
	store = sessions.NewCookieStore([]byte("test"))
	sm := &core.SessionManager{Name: sessionName, Store: store}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := sm.GetSession(r); err != nil {
			sm.SessionErrorRedirect(w, r, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionName, Value: "bad"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusFound {
		t.Fatalf("expected redirect got %d", rr.Code)
	}
	sc := rr.Header().Get("Set-Cookie")
	if !strings.Contains(sc, "Max-Age=0") {
		t.Errorf("expected cleared cookie, got %q", sc)
	}
}

func TestGetSessionOrFailBadSession(t *testing.T) {
	store = sessions.NewCookieStore([]byte("test"))
	sm := &core.SessionManager{Name: sessionName, Store: store}
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionName, Value: "bad"})
	rr := httptest.NewRecorder()
	sess, ok := sm.GetSessionOrFail(rr, req)
	if ok {
		t.Fatalf("expected failure, got session %v", sess)
	}
	if rr.Code != http.StatusFound {
		t.Fatalf("expected redirect got %d", rr.Code)
	}
	sc := rr.Header().Get("Set-Cookie")
	if !strings.Contains(sc, "Max-Age=0") {
		t.Errorf("expected cleared cookie, got %q", sc)
	}
	loc := rr.Header().Get("Location")
	if loc != "/login" {
		t.Errorf("unexpected location %q", loc)
	}
}
