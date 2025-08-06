package linker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

func TestAdminCategoryEditPage(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)

	rows := sqlmock.NewRows([]string{"idlinkercategory", "position", "title", "sortorder"}).
		AddRow(1, 0, "t", 0)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idlinkercategory, position, title, sortorder FROM linker_category WHERE idlinkerCategory = ?")).
		WithArgs(1).WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/admin/linker/categories/category/1/edit", nil)
	req = mux.SetURLVars(req, map[string]string{"category": "1"})
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	AdminCategoryEditPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}
