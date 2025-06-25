package goa4web

import (
	"context"
	"net/http/httptest"
	"testing"

	auth "github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/handlers/common"
	"github.com/gorilla/mux"
)

func TestRequiredAccessAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	ctx := context.WithValue(req.Context(), common.KeyUser, &User{Idusers: 1})
	ctx = context.WithValue(ctx, common.KeyCoreData, &CoreData{SecurityLevel: "writer"})
	req = req.WithContext(ctx)

	if !auth.RequiredAccess("writer")(req, &mux.RouteMatch{}) {
		t.Errorf("expected access allowed")
	}
}

func TestRequiredAccessDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	ctx := context.WithValue(req.Context(), common.KeyUser, &User{Idusers: 1})
	ctx = context.WithValue(ctx, common.KeyCoreData, &CoreData{SecurityLevel: "reader"})
	req = req.WithContext(ctx)

	if auth.RequiredAccess("writer")(req, &mux.RouteMatch{}) {
		t.Errorf("expected access denied")
	}
}
