package admin

import (
	"context"
	"database/sql"
	"fmt"
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

type adminAccessQueries struct {
	db.Querier
	allow bool
}

func (q adminAccessQueries) SystemCheckGrant(_ context.Context, arg db.SystemCheckGrantParams) (int32, error) {
	if arg.Section == common.AdminAccessSection && arg.Action == common.AdminAccessAction {
		if q.allow {
			return 1, nil
		}
		return 0, fmt.Errorf("no admin access grant")
	}
	return 0, fmt.Errorf("unexpected grant check: %#v", arg)
}

func (q adminAccessQueries) SystemCheckRoleGrant(context.Context, db.SystemCheckRoleGrantParams) (int32, error) {
	return 0, sql.ErrNoRows
}

func TestAdminReloadConfigPage_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("POST", "/admin/reload", nil)
	cd := common.NewCoreData(req.Context(), adminAccessQueries{}, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	h := New()
	rr := httptest.NewRecorder()

	h.AdminReloadConfigPage(rr, req)

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
	cd := common.NewCoreData(req.Context(), adminAccessQueries{}, cfg, common.WithUserRoles([]string{"anyone"}))
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
	cd := common.NewCoreData(req.Context(), adminAccessQueries{allow: true}, cfg, common.WithUserRoles([]string{}))
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
	cd := common.NewCoreData(req.Context(), adminAccessQueries{}, cfg, common.WithUserRoles([]string{"anyone"}))
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
	cd := common.NewCoreData(req.Context(), adminAccessQueries{allow: true}, cfg, common.WithUserRoles([]string{}))
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
	cd := common.NewCoreData(req.Context(), adminAccessQueries{allow: true}, cfg, common.WithUserRoles([]string{}))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	h := New()
	if !h.NewServerShutdownTask().Matcher()(req, &mux.RouteMatch{}) {
		t.Fatal("expected matcher success")
	}
}
