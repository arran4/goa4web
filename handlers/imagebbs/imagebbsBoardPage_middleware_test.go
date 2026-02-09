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
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestForbiddenBoardPage(t *testing.T) {
	qs := testhelpers.NewQuerierStub()

	req := httptest.NewRequest("GET", "/imagebbs/board/3", nil)
	req = mux.SetURLVars(req, map[string]string{"board": "3"})

	cd := common.NewCoreData(req.Context(), qs, config.NewRuntimeConfig())

	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	ImagebbsBoardPage(rr, req)

	// Direct call should NOT be forbidden anymore (it bypasses middleware)
	if rr.Code == http.StatusForbidden {
		t.Errorf("expected NOT 403 Forbidden, got %d", rr.Code)
	}
}

func TestForbiddenBoardPageWithMiddleware(t *testing.T) {
	qs := testhelpers.NewQuerierStub()

	req := httptest.NewRequest("GET", "/imagebbs/board/3", nil)
	// We set boardno because the middleware expects it.
	// We set board because the handler might use it (though routes.go uses boardno, handler logic has fallback)
	req = mux.SetURLVars(req, map[string]string{"boardno": "3", "board": "3"})

	cd := common.NewCoreData(req.Context(), qs, config.NewRuntimeConfig())

	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	// Wrap with middleware as done in routes.go
	h := handlers.EnforceGrantFromPath(ImagebbsBoardPage, "imagebbs", "board", "view", "boardno")
	h(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d", rr.Code)
	}
}
