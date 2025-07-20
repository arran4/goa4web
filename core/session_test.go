package core_test

import (
	"context"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

var (
	store       *sessions.CookieStore
	sessionName = "test-session"
)

func TestSessionMiddlewareBadSession(t *testing.T) {
	store = sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = sessionName
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := core.GetSession(r); err != nil {
			core.SessionErrorRedirect(w, r, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionName, Value: "bad"})
	ctx := context.WithValue(req.Context(), consts.KeyQueries, dbpkg.New(nil))
	req = req.WithContext(ctx)
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
	core.Store = store
	core.SessionName = sessionName
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionName, Value: "bad"})
	rr := httptest.NewRecorder()
	sess, ok := core.GetSessionOrFail(rr, req)
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
