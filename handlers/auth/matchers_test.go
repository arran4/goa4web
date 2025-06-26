package auth

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

func TestRequiredAccessAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	ctx := context.WithValue(req.Context(), common.KeyUser, &db.User{Idusers: 1})
	ctx = context.WithValue(ctx, common.KeyCoreData, &common.CoreData{SecurityLevel: "writer"})
	req = req.WithContext(ctx)

	if !RequiredAccess("writer")(req, &mux.RouteMatch{}) {
		t.Errorf("expected access allowed")
	}
}

func TestRequiredAccessDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	ctx := context.WithValue(req.Context(), common.KeyUser, &db.User{Idusers: 1})
	ctx = context.WithValue(ctx, common.KeyCoreData, &common.CoreData{SecurityLevel: "reader"})
	req = req.WithContext(ctx)

	if RequiredAccess("writer")(req, &mux.RouteMatch{}) {
		t.Errorf("expected access denied")
	}
}
