package admin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	adminapi "github.com/arran4/goa4web/internal/adminapi"
	serverpkg "github.com/arran4/goa4web/internal/app/server"
)

func TestAdminAPIServerShutdown_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("POST", "/admin/api/shutdown", nil)
	rr := httptest.NewRecorder()
	AdminAPIServerShutdown(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want %d got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestAdminAPIServerShutdown_Authorized(t *testing.T) {
	adminapi.SetSigningKey("k")
	Srv = &serverpkg.Server{}
	ts, sig := adminapi.Sign("POST", "/admin/api/shutdown")
	req := httptest.NewRequest("POST", "/admin/api/shutdown", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Goa4web %d:%s", ts, sig))
	rr := httptest.NewRecorder()
	AdminAPIServerShutdown(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("want %d got %d", http.StatusOK, rr.Code)
	}
}
