package goa4web

import (
	"context"
	"github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/middleware"
	routerpkg "github.com/arran4/goa4web/internal/router"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoleCheckerMiddlewareAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/admin", nil)
	ctx := context.WithValue(req.Context(), common.KeyCoreData, &CoreData{SecurityLevel: "administrator"})
	req = req.WithContext(ctx)

	called := false
	h := middleware.NewMiddlewareChain(
		routerpkg.RoleCheckerMiddleware("administrator"),
	).Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if !called {
		t.Errorf("handler not called")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("unexpected status %d", rr.Code)
	}
}

func TestRoleCheckerMiddlewareDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/admin", nil)
	ctx := context.WithValue(req.Context(), common.KeyCoreData, &CoreData{SecurityLevel: "reader"})
	req = req.WithContext(ctx)

	called := false
	h := middleware.NewMiddlewareChain(
		routerpkg.RoleCheckerMiddleware("administrator"),
	).Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if called {
		t.Errorf("handler should not be called")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected template render, got status %d", rr.Code)
	}
}
