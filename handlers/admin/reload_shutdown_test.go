package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/app/server"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestAdminReload(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		r := mux.NewRouter()
		ar := r.PathPrefix("/admin").Subrouter()
		cfg := config.NewRuntimeConfig()
		h := New(WithServer(&server.Server{Config: &config.RuntimeConfig{}}))
		navReg := navigation.NewRegistry()
		h.RegisterRoutes(ar, cfg, navReg)
		q := testhelpers.NewQuerierStub()

		req := httptest.NewRequest("POST", "/admin/reload", nil)
		cd := common.NewCoreData(req.Context(), q, cfg,
			common.WithUserRoles([]string{"administrator"}),
			common.WithPermissions([]*db.GetPermissionsByUserIDRow{
				{Name: "administrator", IsAdmin: true},
			}),
		)
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected %d got %d", http.StatusOK, rr.Code)
		}
	})

	t.Run("Unauthorized Page", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		req := httptest.NewRequest("POST", "/admin/reload", nil)
		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		h := New()
		rr := httptest.NewRecorder()

		h.AdminReloadConfigPage(rr, req)

		if rr.Result().StatusCode != http.StatusForbidden {
			t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
		}
	})

	t.Run("Unauthorized Route", func(t *testing.T) {
		r := mux.NewRouter()
		ar := r.PathPrefix("/admin").Subrouter()
		cfg := config.NewRuntimeConfig()
		h := New(WithServer(&server.Server{Config: &config.RuntimeConfig{}}))
		navReg := navigation.NewRegistry()
		h.RegisterRoutes(ar, cfg, navReg)
		q := testhelpers.NewQuerierStub()

		req := httptest.NewRequest("POST", "/admin/reload", nil)
		cd := common.NewCoreData(req.Context(), q, cfg, common.WithUserRoles([]string{"anyone"}))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Code)
		}
	})
}

func TestAdminShutdown(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		h := New(WithServer(&server.Server{}))
		r := mux.NewRouter()
		ar := r.PathPrefix("/admin").Subrouter()
		cfg := config.NewRuntimeConfig()
		navReg := navigation.NewRegistry()
		h.RegisterRoutes(ar, cfg, navReg)
		q := testhelpers.NewQuerierStub()

		form := url.Values{}
		form.Set("task", string(TaskServerShutdown))
		req := httptest.NewRequest("POST", "/admin/shutdown", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		cd := common.NewCoreData(req.Context(), q, cfg,
			common.WithUserRoles([]string{"administrator"}),
			common.WithPermissions([]*db.GetPermissionsByUserIDRow{
				{Name: "administrator", IsAdmin: true},
			}),
		)
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected %d got %d", http.StatusOK, rr.Code)
		}
	})

	t.Run("Unauthorized Route", func(t *testing.T) {
		r := mux.NewRouter()
		ar := r.PathPrefix("/admin").Subrouter()
		cfg := config.NewRuntimeConfig()
		h := New(WithServer(&server.Server{}))
		navReg := navigation.NewRegistry()
		h.RegisterRoutes(ar, cfg, navReg)
		q := testhelpers.NewQuerierStub()

		req := httptest.NewRequest("POST", "/admin/shutdown", nil)
		cd := common.NewCoreData(req.Context(), q, cfg, common.WithUserRoles([]string{"anyone"}))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected %d got %d", http.StatusNotFound, rr.Code)
		}
	})

	t.Run("Task Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/admin/shutdown", nil)
		cfg := config.NewRuntimeConfig()
		q := testhelpers.NewQuerierStub()
		cd := common.NewCoreData(req.Context(), q, cfg, common.WithUserRoles([]string{"anyone"}))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		h := New()
		rr := httptest.NewRecorder()

		handlers.TaskHandler(h.NewServerShutdownTask())(rr, req)

		if rr.Result().StatusCode != http.StatusForbidden {
			t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
		}
	})

	t.Run("Matcher Denied", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/admin/shutdown", nil)
		cfg := config.NewRuntimeConfig()
		q := testhelpers.NewQuerierStub()
		cd := common.NewCoreData(req.Context(), q, cfg, common.WithUserRoles([]string{"anyone"}))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		h := New()
		if h.NewServerShutdownTask().Matcher()(req, &mux.RouteMatch{}) {
			t.Fatal("expected matcher failure")
		}
	})

	t.Run("Matcher Allowed", func(t *testing.T) {
		body := strings.NewReader("task=" + url.QueryEscape(string(TaskServerShutdown)))
		req := httptest.NewRequest(http.MethodPost, "/admin/shutdown", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		cfg := config.NewRuntimeConfig()
		q := testhelpers.NewQuerierStub()
		cd := common.NewCoreData(req.Context(), q, cfg,
			common.WithUserRoles([]string{"administrator"}),
			common.WithPermissions([]*db.GetPermissionsByUserIDRow{
				{Name: "administrator", IsAdmin: true},
			}),
		)
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		h := New()
		if !h.NewServerShutdownTask().Matcher()(req, &mux.RouteMatch{}) {
			t.Fatal("expected matcher success")
		}
	})
}
