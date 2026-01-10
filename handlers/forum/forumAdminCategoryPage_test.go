package forum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func setupRequest(t *testing.T, queries db.Querier, path string, vars map[string]string) (*http.Request, *httptest.ResponseRecorder) {
	t.Helper()
	req := httptest.NewRequest("GET", path, nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, vars)
	rr := httptest.NewRecorder()
	return req, rr
}

func TestAdminCategoryPageLinks(t *testing.T) {
	queries := &db.QuerierStub{
		GetForumCategoryByIdReturns: &db.Forumcategory{
			Idforumcategory:              1,
			ForumcategoryIdforumcategory: 0,
			LanguageID:                   sql.NullInt32{Int32: 0, Valid: true},
			Title:                        sql.NullString{String: "cat", Valid: true},
			Description:                  sql.NullString{String: "desc", Valid: true},
		},
		GetAllForumTopicsByCategoryIdForUserWithLastPosterNameReturns: []*db.GetAllForumTopicsByCategoryIdForUserWithLastPosterNameRow{
			{
				Idforumtopic:                 1,
				ForumcategoryIdforumcategory: 1,
				Title:                        sql.NullString{String: "t", Valid: true},
				Description:                  sql.NullString{String: "d", Valid: true},
				Threads:                      sql.NullInt32{Int32: 0, Valid: true},
				Comments:                     sql.NullInt32{Int32: 0, Valid: true},
				Lastaddition:                 sql.NullTime{Time: time.Now(), Valid: true},
				Handler:                      "",
			},
		},
	}

	req, rr := setupRequest(t, queries, "/admin/forum/categories/category/1", map[string]string{"category": "1"})

	AdminCategoryPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "/admin/forum/categories/category/1/edit") {
		t.Fatalf("missing edit link")
	}
	if !strings.Contains(body, "/admin/forum/categories/category/1/grants") {
		t.Fatalf("missing grants link")
	}
	if !strings.Contains(body, "<a href=\"/admin/forum/topic/1\">1</a>") {
		t.Fatalf("missing topic link")
	}
}

func TestAdminCategoryEditPage(t *testing.T) {
	queries := &db.QuerierStub{
		GetForumCategoryByIdReturns: &db.Forumcategory{
			Idforumcategory:              1,
			ForumcategoryIdforumcategory: 0,
			LanguageID:                   sql.NullInt32{Int32: 0, Valid: true},
			Title:                        sql.NullString{String: "cat", Valid: true},
			Description:                  sql.NullString{String: "desc", Valid: true},
		},
		GetAllForumCategoriesReturns: []*db.Forumcategory{
			{
				Idforumcategory:              1,
				ForumcategoryIdforumcategory: 0,
				LanguageID:                   sql.NullInt32{Int32: 0, Valid: true},
				Title:                        sql.NullString{String: "cat", Valid: true},
				Description:                  sql.NullString{String: "desc", Valid: true},
			},
		},
	}

	req, rr := setupRequest(t, queries, "/admin/forum/categories/category/1/edit", map[string]string{"category": "1"})

	AdminCategoryEditPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestAdminCategoryGrantsPage(t *testing.T) {
	queries := &db.QuerierStub{
		AdminListRolesReturns: []*db.Role{
			{ID: 1, Name: "user", CanLogin: true},
		},
		ListGrantsReturns: []*db.Grant{
			{
				ID:      1,
				RoleID:  sql.NullInt32{Int32: 1, Valid: true},
				Section: "forum",
				Item:    sql.NullString{String: "category", Valid: true},
				ItemID:  sql.NullInt32{Int32: 1, Valid: true},
				Action:  "see",
				Active:  true,
			},
		},
	}

	req, rr := setupRequest(t, queries, "/admin/forum/categories/category/1/grants", map[string]string{"category": "1"})

	AdminCategoryGrantsPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}
