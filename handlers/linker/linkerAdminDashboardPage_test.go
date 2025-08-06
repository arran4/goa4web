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
)

func TestAdminDashboardPage(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)

	rows := sqlmock.NewRows([]string{"idlinkerCategory", "title", "position", "Linkcount"}).
		AddRow(1, "a", 0, 2)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT c.idlinkerCategory, c.title, c.position, COUNT(l.idlinker) as LinkCount\nFROM linker_category c\nLEFT JOIN linker l ON c.idlinkerCategory = l.linker_category_id AND l.listed IS NOT NULL AND l.deleted_at IS NULL\nGROUP BY c.idlinkerCategory\nORDER BY c.position")).WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/admin/linker", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	AdminDashboardPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}
