package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"database/sql"

	"github.com/gorilla/mux"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestRequireGrantForPathInt_MuxRegression(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(params db.SystemCheckGrantParams) (int32, error) {
		return 1, nil // simulate grant allowed
	}
	cfg := &config.RuntimeConfig{}
	cd := common.NewCoreData(context.Background(), q, cfg)

	matcher := RequireGrantForPathInt("news", "post", "view", "id")

	r := mux.NewRouter()
	called := false
	r.Handle("/test/{id}", matcher(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))).Methods("GET")

	req := httptest.NewRequest("GET", "/test/123", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if !called {
		t.Errorf("Handler should have been called")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestRequireGrantForPathInt_MuxRegression_Deny(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(params db.SystemCheckGrantParams) (int32, error) {
		return 0, sql.ErrNoRows // simulate grant denied
	}
	cfg := &config.RuntimeConfig{}
	cd := common.NewCoreData(context.Background(), q, cfg)

	matcher := RequireGrantForPathInt("news", "post", "view", "id")

	r := mux.NewRouter()
	called := false
	r.Handle("/test/{id}", matcher(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))).Methods("GET")

	req := httptest.NewRequest("GET", "/test/123", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if called {
		t.Errorf("Handler should NOT have been called")
	}
	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}
}
