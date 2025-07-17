package auth

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

func TestRequiredAccessAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	cd := corecommon.NewCoreData(req.Context(), nil)
	cd.UserID = 1
	cd.SetRoles([]string{"content writer"})
	ctx := context.WithValue(req.Context(), handlers.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if !RequiredAccess("content writer")(req, &mux.RouteMatch{}) {
		t.Errorf("expected access allowed")
	}
}

func TestRequiredAccessDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	cd := corecommon.NewCoreData(req.Context(), nil)
	cd.UserID = 1
	cd.SetRoles([]string{"anonymous"})
	ctx := context.WithValue(req.Context(), handlers.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if RequiredAccess("content writer")(req, &mux.RouteMatch{}) {
		t.Errorf("expected access denied")
	}
}
