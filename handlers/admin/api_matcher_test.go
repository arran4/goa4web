package admin

import (
	"fmt"
	"net/http/httptest"
	"testing"

	adminapi "github.com/arran4/goa4web/internal/adminapi"
	"github.com/gorilla/mux"
)

func TestAdminAPISigned(t *testing.T) {
	adminapi.SetSigningKey("k")
	ts, sig := adminapi.Sign("POST", "/admin/api/shutdown")
	req := httptest.NewRequest("POST", "/admin/api/shutdown", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Goa4web %d:%s", ts, sig))
	if !AdminAPISigned()(req, &mux.RouteMatch{}) {
		t.Fatal("expected matcher success")
	}
}

func TestAdminAPISignedFail(t *testing.T) {
	adminapi.SetSigningKey("k")
	ts, sig := adminapi.Sign("POST", "/admin/api/shutdown")
	req := httptest.NewRequest("POST", "/admin/api/shutdown", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Goa4web %d:%s", ts, sig))
	req.Method = "GET"
	if AdminAPISigned()(req, &mux.RouteMatch{}) {
		t.Fatal("expected matcher failure")
	}
}
