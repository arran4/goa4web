package blogs

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestHappyPathBloggerListPageSearchRedirect(t *testing.T) {
	// Setup
	queries := testhelpers.NewQuerierStub()
	queries.ListBloggersSearchForListerReturns = []*db.ListBloggersSearchForListerRow{
		{Username: sql.NullString{String: "arran4", Valid: true}, Count: 2},
	}

	req := httptest.NewRequest("GET", "/blogs/bloggers?search=arran4", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	cd.ShareSignKey = "secret"
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	// Execution
	BloggerListPage(rr, req)

	// Verification
	if rr.Result().StatusCode != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
	if loc := rr.Result().Header.Get("Location"); loc != "/blogs/blogger/arran4" {
		t.Fatalf("location=%s", loc)
	}
}
