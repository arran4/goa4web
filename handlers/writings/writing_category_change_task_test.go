package writings

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestWritingCategoryChangeTask(t *testing.T) {
	t.Run("Happy Path - Success", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.SystemListWritingCategoriesReturns = []*db.WritingCategory{
			{Idwritingcategory: 1, Title: sql.NullString{String: "a", Valid: true}},
		}
		q.AdminUpdateWritingCategoryFn = func(ctx context.Context, arg db.AdminUpdateWritingCategoryParams) error {
			if arg.Idwritingcategory != 1 {
				t.Errorf("expected update id 1, got %d", arg.Idwritingcategory)
			}
			return nil
		}

		form := url.Values{"name": {"A"}, "desc": {"B"}, "pcid": {"0"}, "cid": {"1"}}
		req := httptest.NewRequest("POST", "/admin/writings/categories/category/1/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		if v := writingCategoryChangeTask.Action(nil, req); v != nil {
			t.Fatalf("action returned %v", v)
		}
		if len(q.AdminUpdateWritingCategoryCalls) != 1 {
			t.Fatalf("expected 1 update call, got %d", len(q.AdminUpdateWritingCategoryCalls))
		}
	})

	t.Run("Unhappy Path - Loop Detection", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		// 1 -> 2 -> 1 (loop if we try to set 1's parent to 2)
		q.SystemListWritingCategoriesReturns = []*db.WritingCategory{
			{Idwritingcategory: 1, WritingCategoryID: sql.NullInt32{Int32: 2, Valid: true}, Title: sql.NullString{String: "a", Valid: true}},
			{Idwritingcategory: 2, WritingCategoryID: sql.NullInt32{Int32: 1, Valid: true}, Title: sql.NullString{String: "b", Valid: true}},
		}

		// Try to set 1's parent to 2 (which is already 1's child effectively if 2 points to 1... wait)
		// If 1 is parent of 2. And we try to set 1's parent to 2. That creates a loop: 1->2->1.
		// In the setup above:
		// 1 has parent 2.
		// 2 has parent 1.
		// Use a simpler setup.
		// Existing: 1 -> 0.
		// We want to update 1 to have parent 1. (Self loop)
		// Or 1 -> 2 -> 1.

		// Let's use the test case from the original file:
		// rows := sqlmock.NewRows([]string{"idwritingcategory", "writing_category_id", "title", "description"}).
		// 	AddRow(1, 2, "a", "").
		// 	AddRow(2, 1, "b", "")
		// form := url.Values{"name": {"A"}, "desc": {"B"}, "pcid": {"2"}, "cid": {"1"}}

		q.SystemListWritingCategoriesReturns = []*db.WritingCategory{
			{Idwritingcategory: 1, WritingCategoryID: sql.NullInt32{Int32: 2, Valid: true}, Title: sql.NullString{String: "a", Valid: true}},
			{Idwritingcategory: 2, WritingCategoryID: sql.NullInt32{Int32: 1, Valid: true}, Title: sql.NullString{String: "b", Valid: true}},
		}

		form := url.Values{"name": {"A"}, "desc": {"B"}, "pcid": {"2"}, "cid": {"1"}}
		req := httptest.NewRequest("POST", "/admin/writings/categories/category/1/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		if v := writingCategoryChangeTask.Action(nil, req); v == nil {
			t.Fatalf("expected error")
		} else if ue, ok := v.(common.UserError); !ok {
			t.Fatalf("expected user error got %T", v)
		} else if !strings.HasPrefix(ue.UserErrorMessage(), "invalid parent category: loop") {
			t.Fatalf("unexpected error message %q", ue.UserErrorMessage())
		}
	})
}
