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
	"github.com/gorilla/mux"
)

func TestAdminCategoryEditPage(t *testing.T) {
	t.Run("Happy Path - Success", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		category := &db.WritingCategory{
			Idwritingcategory: 1,
			WritingCategoryID: sql.NullInt32{Int32: 0, Valid: true},
			Title:             sql.NullString{String: "a", Valid: true},
			Description:       sql.NullString{String: "b", Valid: true},
		}
		queries.GetWritingCategoryByIdRow = category
		queries.ListWritingCategoriesForListerReturns = []*db.WritingCategory{category}
		queries.SystemListWritingCategoriesReturns = []*db.WritingCategory{category}

		req := httptest.NewRequest("GET", "/admin/writings/categories/category/1/edit", nil)
		ctx := req.Context()
		cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		req = mux.SetURLVars(req, map[string]string{"category": "1"})
		rr := httptest.NewRecorder()

		AdminCategoryEditPage(rr, req)

		if rr.Result().StatusCode != http.StatusOK {
			t.Fatalf("status=%d", rr.Result().StatusCode)
		}
	})
}
