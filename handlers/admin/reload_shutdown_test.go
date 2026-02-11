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
)

func TestAdminReloadConfigPage_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("POST", "/admin/reload", nil)
	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	h := New()
	rr := httptest.NewRecorder()

	(&AdminReloadConfigPage{
		ConfigFile: h.ConfigFile,
		Srv:        h.Srv,
		DBPool:     h.DBPool,
	}).ServeHTTP(rr, req)

	if rr.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
	}
}

func TestAdminReloadRoute_Unauthorized(t *testing.T) {
	r := mux.NewRouter()
	ar := r.PathPrefix("/admin").Subrouter()
	cfg := config.NewRuntimeConfig()
	h := New(WithServer(&server.Server{Config: &config.RuntimeConfig{}}))
	navReg := navigation.NewRegistry()
	h.RegisterRoutes(ar, cfg, navReg)

	req := httptest.NewRequest("POST", "/admin/reload", nil)
	cd := common.NewCoreData(req.Context(), nil, cfg, common.WithUserRoles([]string{"anyone"}))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected %d got %d", http.StatusNotFound, rr.Code)
	}
}

func TestAdminReloadRoute_Authorized(t *testing.T) {
	r := mux.NewRouter()
	ar := r.PathPrefix("/admin").Subrouter()
	cfg := config.NewRuntimeConfig()
	h := New(WithServer(&server.Server{Config: &config.RuntimeConfig{}}))
	navReg := navigation.NewRegistry()
	h.RegisterRoutes(ar, cfg, navReg)

	req := httptest.NewRequest("POST", "/admin/reload", nil)
	cd := common.NewCoreData(req.Context(), nil, cfg,
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
}

func TestAdminShutdownRoute_Unauthorized(t *testing.T) {
	r := mux.NewRouter()
	ar := r.PathPrefix("/admin").Subrouter()
	cfg := config.NewRuntimeConfig()
	h := New(WithServer(&server.Server{}))
	navReg := navigation.NewRegistry()
	h.RegisterRoutes(ar, cfg, navReg)

	req := httptest.NewRequest("POST", "/admin/shutdown", nil)
	cd := common.NewCoreData(req.Context(), nil, cfg, common.WithUserRoles([]string{"anyone"}))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected %d got %d", http.StatusNotFound, rr.Code)
	}
}

func TestAdminShutdownRoute_Authorized(t *testing.T) {
	h := New(WithServer(&server.Server{}))
	r := mux.NewRouter()
	ar := r.PathPrefix("/admin").Subrouter()
	cfg := config.NewRuntimeConfig()
	navReg := navigation.NewRegistry()
	h.RegisterRoutes(ar, cfg, navReg)

	form := url.Values{}
	form.Set("task", string(TaskServerShutdown))
	req := httptest.NewRequest("POST", "/admin/shutdown", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cd := common.NewCoreData(req.Context(), nil, cfg,
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
}

func TestServerShutdownTask_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("POST", "/admin/shutdown", nil)
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), nil, cfg, common.WithUserRoles([]string{"anyone"}))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	h := New()
	rr := httptest.NewRecorder()

	handlers.TaskHandler(h.NewServerShutdownTask())(rr, req)

	if rr.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
	}
}

func TestServerShutdownMatcher_Denied(t *testing.T) {
	req := httptest.NewRequest("POST", "/admin/shutdown", nil)
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), nil, cfg, common.WithUserRoles([]string{"anyone"}))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	h := New()
	if h.NewServerShutdownTask().Matcher()(req, &mux.RouteMatch{}) {
		t.Fatal("expected matcher failure")
	}
}

func TestServerShutdownMatcher_Allowed(t *testing.T) {
	body := strings.NewReader("task=" + url.QueryEscape(string(TaskServerShutdown)))
	req := httptest.NewRequest(http.MethodPost, "/admin/shutdown", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), nil, cfg,
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
}
