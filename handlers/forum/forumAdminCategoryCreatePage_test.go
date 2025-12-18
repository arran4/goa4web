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

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminCategoryCreateSubmitSuccess(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	mock.MatchExpectationsInOrder(false)
	categoriesRows := sqlmock.NewRows([]string{"idforumcategory", "forumcategory_idforumcategory", "language_id", "title", "description"})
	mock.ExpectQuery(regexp.QuoteMeta("SELECT f.idforumcategory, f.forumcategory_idforumcategory, f.language_id, f.title, f.description\nFROM forumcategory f\nWHERE (\n    f.language_id = 0\n    OR f.language_id IS NULL\n    OR EXISTS (\n        SELECT 1 FROM user_language ul\n        WHERE ul.users_idusers = ?\n          AND ul.language_id = f.language_id\n    )\n    OR NOT EXISTS (\n        SELECT 1 FROM user_language ul WHERE ul.users_idusers = ?\n    )\n)\n")).WithArgs(int32(0), int32(0)).WillReturnRows(categoriesRows)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO forumcategory (forumcategory_idforumcategory, language_id, title, description)\nVALUES (?, ?, ?, ?)")).WithArgs(
		int32(1),
		sql.NullInt32{Int32: 2, Valid: true},
		sql.NullString{String: "name", Valid: true},
		sql.NullString{String: "desc", Valid: true},
	).WillReturnResult(sqlmock.NewResult(5, 1))

	queries := db.New(sqlDB)
	form := url.Values{
		"name":     {"name"},
		"desc":     {"desc"},
		"pcid":     {"1"},
		"language": {"2"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/forum/categories/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	AdminCategoryCreateSubmit(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	location := rr.Header().Get("Location")
	if !strings.Contains(location, "/admin/forum/categories") || !strings.Contains(location, "error=category created") {
		t.Fatalf("unexpected redirect location %q", location)
	}
}

func TestAdminCategoryCreateSubmitValidationError(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)
	form := url.Values{
		"desc":     {"desc"},
		"pcid":     {"1"},
		"language": {"2"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/forum/categories/create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	AdminCategoryCreateSubmit(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	location := rr.Header().Get("Location")
	if !strings.Contains(location, "/admin/forum/categories/create") || !strings.Contains(location, "error=missing name") {
		t.Fatalf("expected validation error in redirect, got %q", location)
	}
}
