package auth

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers/common"
)

func TestRequiredAccessAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	cd := corecommon.NewCoreData(req.Context(), nil)
	cd.UserID = 1
	cd.SetRole("writer")
	ctx := context.WithValue(req.Context(), common.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if !RequiredAccess("writer")(req, &mux.RouteMatch{}) {
		t.Errorf("expected access allowed")
	}
}

func TestRequiredAccessDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	cd := corecommon.NewCoreData(req.Context(), nil)
	cd.UserID = 1
	cd.SetRole("reader")
	ctx := context.WithValue(req.Context(), common.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if RequiredAccess("writer")(req, &mux.RouteMatch{}) {
		t.Errorf("expected access denied")
	}
}
