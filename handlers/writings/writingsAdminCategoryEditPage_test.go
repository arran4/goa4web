package writings

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

	catRows := sqlmock.NewRows([]string{"idwritingcategory", "writing_category_id", "title", "description"}).
		AddRow(1, 0, "a", "b")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idwritingcategory, writing_category_id, title, description FROM writing_category WHERE idwritingCategory = ?")).
		WithArgs(int32(1)).WillReturnRows(catRows)

	listRows := sqlmock.NewRows([]string{"idwritingcategory", "writing_category_id", "title", "description"}).
		AddRow(1, 0, "a", "b")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT wc.idwritingcategory, wc.writing_category_id, wc.title, wc.description\nFROM writing_category wc")).
		WillReturnRows(listRows)

	req := httptest.NewRequest("GET", "/admin/writings/categories/category/1/edit", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"category": "1"})
	rr := httptest.NewRecorder()

	AdminCategoryEditPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}
