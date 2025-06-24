package goa4web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
)

func TestCoreAdderMiddlewareBadSession(t *testing.T) {
	store = sessions.NewCookieStore([]byte("test"))
	h := CoreAdderMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionName, Value: "bad"})
	ctx := context.WithValue(req.Context(), ContextValues("queries"), New(nil))
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
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionName, Value: "bad"})
	rr := httptest.NewRecorder()
	sess, ok := GetSessionOrFail(rr, req)
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
