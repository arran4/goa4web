package forum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func setupCategoryEditRequest(t *testing.T, queries db.Querier, path string, form url.Values, vars map[string]string) (*http.Request, *httptest.ResponseRecorder) {
	t.Helper()
	req := httptest.NewRequest("POST", path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, vars)
	rr := httptest.NewRecorder()
	return req, rr
}

func TestAdminCategoryEditSubmitSuccess(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.GetForumCategoryByIdReturns = &db.Forumcategory{
		Idforumcategory:              1,
		ForumcategoryIdforumcategory: 0,
		LanguageID:                   sql.NullInt32{Int32: 0, Valid: true},
		Title:                        sql.NullString{String: "cat", Valid: true},
		Description:                  sql.NullString{String: "desc", Valid: true},
	}
	queries.GetAllForumCategoriesReturns = []*db.Forumcategory{
		{
			Idforumcategory:              1,
			ForumcategoryIdforumcategory: 0,
			LanguageID:                   sql.NullInt32{Int32: 0, Valid: true},
			Title:                        sql.NullString{String: "cat", Valid: true},
			Description:                  sql.NullString{String: "desc", Valid: true},
		},
		{
			Idforumcategory:              2,
			ForumcategoryIdforumcategory: 0,
			LanguageID:                   sql.NullInt32{Int32: 0, Valid: true},
			Title:                        sql.NullString{String: "parent", Valid: true},
			Description:                  sql.NullString{String: "pdesc", Valid: true},
		},
	}
	queries.AdminUpdateForumCategoryFn = func(ctx context.Context, arg db.AdminUpdateForumCategoryParams) error {
		if arg.Title.String != "Updated" || !arg.Title.Valid {
			t.Fatalf("unexpected title %+v", arg.Title)
		}
		if arg.Description.String != "Updated desc" || !arg.Description.Valid {
			t.Fatalf("unexpected desc %+v", arg.Description)
		}
		if arg.ParentID != 2 {
			t.Fatalf("unexpected parent %d", arg.ParentID)
		}
		if arg.LanguageID != (sql.NullInt32{Int32: 3, Valid: true}) {
			t.Fatalf("unexpected language %+v", arg.LanguageID)
		}
		if arg.Idforumcategory != 1 {
			t.Fatalf("unexpected category id %d", arg.Idforumcategory)
		}
		return nil
	}

	form := url.Values{
		"name":     {"Updated"},
		"desc":     {"Updated desc"},
		"pcid":     {"2"},
		"language": {"3"},
	}
	req, rr := setupCategoryEditRequest(t, queries, "/admin/forum/categories/category/1/edit", form, map[string]string{"category": "1"})

	AdminCategoryEditSubmit(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/admin/forum/categories/category/1" {
		t.Fatalf("redirect=%s", loc)
	}
}

func TestAdminCategoryEditSubmitMissingCategory(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.GetForumCategoryByIdErr = sql.ErrNoRows

	form := url.Values{
		"name": {"Updated"},
		"desc": {"Updated desc"},
		"pcid": {"0"},
	}
	req, rr := setupCategoryEditRequest(t, queries, "/admin/forum/categories/category/99/edit", form, map[string]string{"category": "99"})

	AdminCategoryEditSubmit(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestAdminCategoryEditSubmitValidationError(t *testing.T) {
	queries := testhelpers.NewQuerierStub()

	form := url.Values{
		"name": {""},
		"desc": {"Updated desc"},
		"pcid": {"1"},
	}
	req, rr := setupCategoryEditRequest(t, queries, "/admin/forum/categories/category/1/edit", form, map[string]string{"category": "1"})

	AdminCategoryEditSubmit(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); !strings.HasSuffix(loc, "?error=category name cannot be empty") {
		t.Fatalf("redirect=%s", loc)
	}
}
