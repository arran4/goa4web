package privateforum

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestStartGroupDiscussionPage_RouterGrantFailure(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/private/topic/new", nil)
	w := httptest.NewRecorder()

	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(params db.SystemCheckGrantParams) (int32, error) {
		return 1, nil // Grant is allowed
	}
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	cd.UserID = 1 // Logged in

	r := mux.NewRouter()
	cfg := config.NewRuntimeConfig()

	RegisterRoutes(r, cfg)

	injectedHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		r.ServeHTTP(w, req)
	})

	injectedHandler.ServeHTTP(w, req)

	t.Logf("Status Code: %d", w.Code)
	if w.Code == http.StatusNotFound {
		t.Fatalf("Expected route to match, but got 404. Matcher failed?")
	}
	if w.Code == http.StatusForbidden {
		t.Fatalf("Expected grant to pass, got 403 Forbidden")
	}
}
