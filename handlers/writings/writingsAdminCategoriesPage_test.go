package writings

import (
	"context"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestWritingsAdminCategoriesPage(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)

	rows := sqlmock.NewRows([]string{"idwritingcategory", "writing_category_id", "title", "description"}).
		AddRow(1, 0, "a", "b")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT wc.idwritingcategory, wc.writing_category_id, wc.title, wc.description\nFROM writing_category wc")).WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/admin/writings/categories", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" }))
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	AdminCategoriesPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}
