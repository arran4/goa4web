package imagebbs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestCheckBoardViewGrant_Denied(t *testing.T) {
	// Setup CoreData with denied grants
	qs := testhelpers.NewQuerierStub(testhelpers.WithGrantResult(false))
	req := httptest.NewRequest("GET", "/imagebbs/board/1", nil)
	req = mux.SetURLVars(req, map[string]string{"board": "1"})

	cd := common.NewCoreData(req.Context(), qs, config.NewRuntimeConfig())
	ctx := req.Context()
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	// Dummy handler that should not be called
	handlerCalled := false
	dummyHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}

	// Wrap with middleware
	wrapped := CheckBoardViewGrant(dummyHandler)

	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if handlerCalled {
		t.Error("Handler should not have been called")
	}
	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 Forbidden, got %d", rr.Code)
	}
}

func TestCheckBoardViewGrant_Allowed(t *testing.T) {
	// Setup CoreData with allowed grants
	qs := testhelpers.NewQuerierStub(testhelpers.WithGrantResult(true))
	req := httptest.NewRequest("GET", "/imagebbs/board/1", nil)
	req = mux.SetURLVars(req, map[string]string{"board": "1"})

	cd := common.NewCoreData(req.Context(), qs, config.NewRuntimeConfig())
	ctx := req.Context()
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	// Dummy handler that SHOULD be called
	handlerCalled := false
	dummyHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}

	// Wrap with middleware
	wrapped := CheckBoardViewGrant(dummyHandler)

	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if !handlerCalled {
		t.Error("Handler should have been called")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", rr.Code)
	}
}
