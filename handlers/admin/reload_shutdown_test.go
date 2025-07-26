package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

func TestAdminReloadConfigPage_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("POST", "/admin/reload", nil)
	cd := common.NewCoreData(req.Context(), nil, config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" }))
	cd.SetRoles([]string{"anonymous"})
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	AdminReloadConfigPage(rr, req)

	if rr.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
	}
}

func TestAdminReloadRoute_Unauthorized(t *testing.T) {
	r := mux.NewRouter()
	ar := r.PathPrefix("/admin").Subrouter()
	cfg := config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	RegisterRoutes(ar, cfg)

	req := httptest.NewRequest("POST", "/admin/reload", nil)
	cd := common.NewCoreData(req.Context(), nil, cfg)
	cd.SetRoles([]string{"anonymous"})
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected %d got %d", http.StatusNotFound, rr.Code)
	}
}

func TestAdminShutdownRoute_Unauthorized(t *testing.T) {
	r := mux.NewRouter()
	ar := r.PathPrefix("/admin").Subrouter()
	cfg := config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	RegisterRoutes(ar, cfg)

	req := httptest.NewRequest("POST", "/admin/shutdown", nil)
	cd := common.NewCoreData(req.Context(), nil, cfg)
	cd.SetRoles([]string{"anonymous"})
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected %d got %d", http.StatusNotFound, rr.Code)
	}
}

func TestServerShutdownTask_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("POST", "/admin/shutdown", nil)
	cfg := config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	cd := common.NewCoreData(req.Context(), nil, cfg)
	cd.SetRoles([]string{"anonymous"})
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handlers.TaskHandler(serverShutdownTask)(rr, req)

	if rr.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
	}
}

func TestServerShutdownMatcher_Denied(t *testing.T) {
	req := httptest.NewRequest("POST", "/admin/shutdown", nil)
	cfg := config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	cd := common.NewCoreData(req.Context(), nil, cfg)
	cd.SetRoles([]string{"anonymous"})
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	if serverShutdownTask.Matcher()(req, &mux.RouteMatch{}) {
		t.Fatal("expected matcher failure")
	}
}

func TestServerShutdownMatcher_Allowed(t *testing.T) {
	req := httptest.NewRequest("POST", "/admin/shutdown", nil)
	req.Form = make(map[string][]string)
	req.Form.Set("task", string(TaskServerShutdown))
	cfg := config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	cd := common.NewCoreData(req.Context(), nil, cfg)
	cd.SetRoles([]string{"administrator"})
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	if !serverShutdownTask.Matcher()(req, &mux.RouteMatch{}) {
		t.Fatal("expected matcher success")
	}
}
