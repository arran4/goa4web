package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
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
	t.Run("Missing ID in vars", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/linker/show/123", nil)
		// No vars set
		cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		h := EnforceGrant(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}, "linker", "link", "view", "link")

		rr := httptest.NewRecorder()
		h(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rr.Code)
		}
	})

	t.Run("Invalid ID (string)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/linker/show/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"link": "abc"})
		cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		h := EnforceGrant(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}, "linker", "link", "view", "link")

		rr := httptest.NewRecorder()
		h(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rr.Code)
		}
	})

	t.Run("Invalid ID (0)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/linker/show/0", nil)
		req = mux.SetURLVars(req, map[string]string{"link": "0"})
		cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		h := EnforceGrant(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}, "linker", "link", "view", "link")

		rr := httptest.NewRecorder()
		h(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rr.Code)
		}
	})

	t.Run("Denied", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/linker/show/123", nil)
		req = mux.SetURLVars(req, map[string]string{"link": "123"})

		q := testhelpers.NewQuerierStub()
		q.SystemCheckGrantFn = func(params db.SystemCheckGrantParams) (int32, error) {
			return 0, errors.New("denied")
		}

		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		h := EnforceGrant(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}, "linker", "link", "view", "link")

		rr := httptest.NewRecorder()
		h(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rr.Code)
		}
	})

	t.Run("Allowed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/linker/show/123", nil)
		req = mux.SetURLVars(req, map[string]string{"link": "123"})

		q := testhelpers.NewQuerierStub()
		q.SystemCheckGrantFn = func(params db.SystemCheckGrantParams) (int32, error) {
			return 1, nil
		}

		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		h := EnforceGrant(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}, "linker", "link", "view", "link")

		rr := httptest.NewRecorder()
		h(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})
}
