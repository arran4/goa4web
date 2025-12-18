package forum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func setupCategoryEditRequest(t *testing.T, queries *db.Queries, path string, form url.Values, vars map[string]string) (*http.Request, *httptest.ResponseRecorder) {
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
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)
	mock.MatchExpectationsInOrder(false)

	categoryRows := sqlmock.NewRows([]string{"idforumcategory", "forumcategory_idforumcategory", "language_id", "title", "description"}).
		AddRow(1, 0, 0, "cat", "desc")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idforumcategory, forumcategory_idforumcategory, language_id, title, description FROM forumcategory")).
		WithArgs(int32(1), int32(0), int32(0)).
		WillReturnRows(categoryRows)

	allRows := sqlmock.NewRows([]string{"idforumcategory", "forumcategory_idforumcategory", "language_id", "title", "description"}).
		AddRow(1, 0, 0, "cat", "desc").
		AddRow(2, 0, 0, "parent", "pdesc")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT f.idforumcategory, f.forumcategory_idforumcategory, f.language_id, f.title, f.description\nFROM forumcategory f")).
		WithArgs(int32(0), int32(0)).
		WillReturnRows(allRows)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE forumcategory\nSET title = ?,\n    description = ?,\n    forumcategory_idforumcategory = ?,\n    language_id = ?\nWHERE idforumcategory = ?")).
		WithArgs(sql.NullString{String: "Updated", Valid: true}, sql.NullString{String: "Updated desc", Valid: true}, int32(2), sql.NullInt32{Int32: 3, Valid: true}, int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	form := url.Values{
		"name":     {"Updated"},
		"desc":     {"Updated desc"},
		"pcid":     {"2"},
		"language": {"3"},
	}
	req, rr := setupCategoryEditRequest(t, queries, "/admin/forum/categories/category/1/edit", form, map[string]string{"category": "1"})

	AdminCategoryEditSubmit(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/admin/forum/categories/category/1" {
		t.Fatalf("redirect=%s", loc)
	}
}

func TestAdminCategoryEditSubmitMissingCategory(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT idforumcategory, forumcategory_idforumcategory, language_id, title, description FROM forumcategory")).
		WithArgs(int32(99), int32(0), int32(0)).
		WillReturnError(sql.ErrNoRows)

	form := url.Values{
		"name": {"Updated"},
		"desc": {"Updated desc"},
		"pcid": {"0"},
	}
	req, rr := setupCategoryEditRequest(t, queries, "/admin/forum/categories/category/99/edit", form, map[string]string{"category": "99"})

	AdminCategoryEditSubmit(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestAdminCategoryEditSubmitValidationError(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)

	form := url.Values{
		"name": {""},
		"desc": {"Updated desc"},
		"pcid": {"1"},
	}
	req, rr := setupCategoryEditRequest(t, queries, "/admin/forum/categories/category/1/edit", form, map[string]string{"category": "1"})

	AdminCategoryEditSubmit(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); !strings.HasSuffix(loc, "?error=category name cannot be empty") {
		t.Fatalf("redirect=%s", loc)
	}
}
