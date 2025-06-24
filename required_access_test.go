package goa4web

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestRequiredAccessAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	ctx := context.WithValue(req.Context(), ContextValues("user"), &User{Idusers: 1})
	ctx = context.WithValue(ctx, ContextValues("coreData"), &CoreData{SecurityLevel: "writer"})
	req = req.WithContext(ctx)

	if !RequiredAccess("writer")(req, &mux.RouteMatch{}) {
		t.Errorf("expected access allowed")
	}
}

func TestRequiredAccessDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	ctx := context.WithValue(req.Context(), ContextValues("user"), &User{Idusers: 1})
	ctx = context.WithValue(ctx, ContextValues("coreData"), &CoreData{SecurityLevel: "reader"})
	req = req.WithContext(ctx)

	if RequiredAccess("writer")(req, &mux.RouteMatch{}) {
		t.Errorf("expected access denied")
	}
}
