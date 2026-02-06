package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

func TestRequireRole(t *testing.T) {
	h := RequireRole(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, fmt.Errorf("administrator role required"), "administrator")

	t.Run("Unhappy Path - Forbidden", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

		rr := httptest.NewRecorder()
		h(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "administrator role required") {
			t.Fatalf("expected body to contain error message, got %q", rr.Body.String())
		}
	})

	t.Run("Happy Path", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
		rr := httptest.NewRecorder()
		h(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected %d got %d", http.StatusOK, rr.Code)
		}
	})
}
