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

func TestBoardThreadPage_Forbidden(t *testing.T) {
	// Configure stub to deny access (SystemCheckGrant returns false)
	qs := testhelpers.NewQuerierStub(testhelpers.WithGrantResult(false))

	// Setup request with variables
	req := httptest.NewRequest("GET", "/imagebbs/board/3/thread/1", nil)
	req = mux.SetURLVars(req, map[string]string{"board": "3", "thread": "1"})

	// Create CoreData with the QuerierStub
	cd := common.NewCoreData(req.Context(), qs, config.NewRuntimeConfig())

	// Inject CoreData into context
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	BoardThreadPage(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusForbidden)
	}
}
