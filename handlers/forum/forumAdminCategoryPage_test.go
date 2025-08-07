package forum

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func setupRequest(t *testing.T, queries *db.Queries, path string, vars map[string]string) (*http.Request, *httptest.ResponseRecorder) {
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
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)
	mock.MatchExpectationsInOrder(false)

	catRows := sqlmock.NewRows([]string{"idforumcategory", "forumcategory_idforumcategory", "language_idlanguage", "title", "description"}).
		AddRow(1, 0, 0, "cat", "desc")
	mock.ExpectQuery("SELECT idforumcategory, forumcategory_idforumcategory, language_idlanguage, title, description FROM forumcategory").
		WillReturnRows(catRows)

	topicsRows := sqlmock.NewRows([]string{"idforumtopic", "lastposter", "forumcategory_idforumcategory", "language_idlanguage", "title", "description", "threads", "comments", "lastaddition", "handler"}).
		AddRow(1, 0, 1, 0, "t", "d", 0, 0, time.Now(), "")
	mock.ExpectQuery("SELECT idforumtopic, lastposter, forumcategory_idforumcategory, language_idlanguage, title, description, threads, comments, lastaddition, handler FROM forumtopic").
		WillReturnRows(topicsRows)

	req, rr := setupRequest(t, queries, "/admin/forum/category/1", map[string]string{"category": "1"})

	AdminCategoryPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "/admin/forum/category/1/edit") {
		t.Fatalf("missing edit link")
	}
	if !strings.Contains(body, "/admin/forum/category/1/grants") {
		t.Fatalf("missing grants link")
	}
	if !strings.Contains(body, "<a href=\"/admin/forum/topic/1\">1</a>") {
		t.Fatalf("missing topic link")
	}
}

func TestAdminCategoryEditPage(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)
	mock.MatchExpectationsInOrder(false)

	catRows := sqlmock.NewRows([]string{"idforumcategory", "forumcategory_idforumcategory", "language_idlanguage", "title", "description"}).
		AddRow(1, 0, 0, "cat", "desc")
	mock.ExpectQuery("SELECT idforumcategory, forumcategory_idforumcategory, language_idlanguage, title, description FROM forumcategory").
		WillReturnRows(catRows)

	allRows := sqlmock.NewRows([]string{"idforumcategory", "forumcategory_idforumcategory", "language_idlanguage", "title", "description"}).
		AddRow(1, 0, 0, "cat", "desc")
	mock.ExpectQuery("SELECT f.idforumcategory, f.forumcategory_idforumcategory, f.language_idlanguage, f.title, f.description\nFROM forumcategory f").
		WillReturnRows(allRows)

	req, rr := setupRequest(t, queries, "/admin/forum/category/1/edit", map[string]string{"category": "1"})

	AdminCategoryEditPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestAdminCategoryGrantsPage(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)
	mock.MatchExpectationsInOrder(false)

	rolesRows := sqlmock.NewRows([]string{"id", "name", "can_login", "is_admin", "public_profile_allowed_at"}).
		AddRow(1, "user", true, false, nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, can_login, is_admin, public_profile_allowed_at FROM roles ORDER BY id")).
		WillReturnRows(rolesRows)

	grantsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "user_id", "role_id", "section", "item", "rule_type", "item_id", "item_rule", "action", "extra", "active"}).
		AddRow(1, nil, nil, nil, nil, "forum", "category", "allow", 1, nil, "see", nil, true)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, created_at, updated_at, user_id, role_id, section, item, rule_type, item_id, item_rule, action, extra, active FROM grants ORDER BY id")).
		WillReturnRows(grantsRows)

	req, rr := setupRequest(t, queries, "/admin/forum/category/1/grants", map[string]string{"category": "1"})

	AdminCategoryGrantsPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}
