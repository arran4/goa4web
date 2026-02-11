package blogs

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestHappyPathBloggersBloggerPage(t *testing.T) {
	// Setup
	queries := testhelpers.NewQuerierStub()
	queries.ListBloggersForListerReturns = []*db.ListBloggersForListerRow{
		{Username: sql.NullString{String: "bob", Valid: true}, Count: 2},
	}

	req := httptest.NewRequest("GET", "/blogs/bloggers/blogger", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	// Execution
	BloggersBloggerPage(rr, req)

	// Verification
	if rr.Result().StatusCode != 200 {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}
