package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
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

func TestEnforceGrant(t *testing.T) {
	h := EnforceGrant(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, "section", "item", "action", "id")

	t.Run("Missing ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		// No vars set
		cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

		rr := httptest.NewRecorder()
		h(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected %d got %d", http.StatusNotFound, rr.Code)
		}
	})

	t.Run("Invalid ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "abc"})
		cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

		rr := httptest.NewRecorder()
		h(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected %d got %d", http.StatusNotFound, rr.Code)
		}
	})

	t.Run("Denied", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test/123", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "123"})
		q := testhelpers.NewQuerierStub()
		q.SystemCheckGrantFn = func(params db.SystemCheckGrantParams) (int32, error) {
			return 0, errors.New("denied")
		}
		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

		rr := httptest.NewRecorder()
		h(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Code)
		}
	})

	t.Run("Allowed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test/123", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "123"})
		q := testhelpers.NewQuerierStub()
		q.SystemCheckGrantFn = func(params db.SystemCheckGrantParams) (int32, error) {
			return 1, nil
		}
		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
		req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

		rr := httptest.NewRecorder()
		h(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected %d got %d", http.StatusOK, rr.Code)
		}
	})
}
