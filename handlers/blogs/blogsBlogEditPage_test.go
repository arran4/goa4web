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
)

func TestBlogEditPage_FailsWhenBlogNotLoaded(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/blog/edit/1", nil)
	queries := testhelpers.NewQuerierStub()
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), queries, cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)

	rr := httptest.NewRecorder()

	BlogEditPage(rr, req.WithContext(ctx))

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status Forbidden (%d), got %d", http.StatusForbidden, rr.Code)
	}
}
