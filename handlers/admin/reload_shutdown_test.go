package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

func TestAdminReloadConfigPage_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("POST", "/admin/reload", nil)
	cd := common.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"anonymous"})
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	AdminReloadConfigPage(rr, req)

	if rr.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
	}
}

func TestAdminShutdownPage_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("POST", "/admin/shutdown", nil)
	cd := common.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"anonymous"})
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	AdminShutdownPage(rr, req)

	if rr.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
	}
}
