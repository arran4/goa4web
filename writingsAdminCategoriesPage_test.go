package goa4web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/handlers/common"
)

func TestWritingsAdminCategoriesPage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := New(db)

	rows := sqlmock.NewRows([]string{"idwritingcategory", "writingcategory_idwritingcategory", "title", "description"}).
		AddRow(1, 0, "a", "b")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT wc.idwritingcategory, wc.writingcategory_idwritingcategory, wc.title, wc.description\nFROM writingCategory wc")).WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/admin/writings/categories", nil)
	ctx := context.WithValue(req.Context(), common.KeyQueries, queries)
	ctx = context.WithValue(ctx, common.KeyCoreData, &CoreData{})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	writingsAdminCategoriesPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}
