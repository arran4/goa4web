package blogs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
)

func TestBlogEditPage_FailsWhenBlogNotLoaded(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/blog/edit/1", nil)
	req = mux.SetURLVars(req, map[string]string{"blog": "1"})
	queries := testhelpers.NewQuerierStub()
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), queries, cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)

	rr := httptest.NewRecorder()

	RequireBlogAuthor(http.HandlerFunc(BlogEditPage)).ServeHTTP(rr, req.WithContext(ctx))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status NotFound (%d), got %d", http.StatusNotFound, rr.Code)
	}
}
