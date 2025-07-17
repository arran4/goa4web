package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/middleware"
)

func TestRoleCheckerMiddlewareAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/admin", nil)
	cd := corecommon.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"administrator"})
	ctx := context.WithValue(req.Context(), corecorecommon.KeyCoreData, cd)
	req = req.WithContext(ctx)

	called := false
	h := middleware.NewMiddlewareChain(
		RoleCheckerMiddleware("administrator"),
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
	cd := corecommon.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"anonymous"})
	ctx := context.WithValue(req.Context(), corecorecommon.KeyCoreData, cd)
	req = req.WithContext(ctx)

	called := false
	h := middleware.NewMiddlewareChain(
		RoleCheckerMiddleware("administrator"),
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
