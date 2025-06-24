package goa4web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoleCheckerMiddlewareAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/admin", nil)
	ctx := context.WithValue(req.Context(), ContextValues("coreData"), &CoreData{SecurityLevel: "administrator"})
	req = req.WithContext(ctx)

	called := false
	h := RoleCheckerMiddleware("administrator")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	ctx := context.WithValue(req.Context(), ContextValues("coreData"), &CoreData{SecurityLevel: "reader"})
	req = req.WithContext(ctx)

	called := false
	h := RoleCheckerMiddleware("administrator")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
