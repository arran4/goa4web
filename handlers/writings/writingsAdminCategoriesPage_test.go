package writings

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

func TestWritingsAdminCategoriesPage(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	category := &db.WritingCategory{
		Idwritingcategory: 1,
		WritingCategoryID: sql.NullInt32{Int32: 0, Valid: true},
		Title:             sql.NullString{String: "a", Valid: true},
		Description:       sql.NullString{String: "b", Valid: true},
	}
	queries.ListWritingCategoriesForListerReturns = []*db.WritingCategory{category}
	queries.SystemListWritingCategoriesReturns = []*db.WritingCategory{category}

	req := httptest.NewRequest("GET", "/admin/writings/categories", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	AdminCategoriesPage(rr, req)

	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}
