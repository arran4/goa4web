package admin

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/internal/adminapi"
	"github.com/gorilla/mux"
)

func TestHappyPathAdminAPISigned(t *testing.T) {
	AdminAPISecret = "k"
	signer := adminapi.NewSigner("k")
	ts, sig := signer.Sign("POST", "/admin/api/shutdown")
	req := httptest.NewRequest("POST", "/admin/api/shutdown", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Goa4web %d:%s", ts, sig))
	if !AdminAPISigned()(req, &mux.RouteMatch{}) {
		t.Fatal("expected matcher success")
	}
}

func TestAdminAPISignedFail(t *testing.T) {
	t.Run("Unhappy Path", func(t *testing.T) {
		AdminAPISecret = "k"
		signer := adminapi.NewSigner("k")
		ts, sig := signer.Sign("POST", "/admin/api/shutdown")
		req := httptest.NewRequest("POST", "/admin/api/shutdown", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Goa4web %d:%s", ts, sig))
		req.Method = "GET"
		if AdminAPISigned()(req, &mux.RouteMatch{}) {
			t.Fatal("expected matcher failure")
		}
	})
}
