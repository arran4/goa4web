package news

import (
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestCustomNewsIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	CustomNewsIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("admin not in admin mode should not see add news")
	}

	cd.AdminMode = true
	CustomNewsIndex(cd, req)
	if !common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("admin should see add news when admin mode is enabled")
	}

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	ctx := req.Context()
	cd = common.NewCoreData(ctx, db.New(conn), config.NewRuntimeConfig(), common.WithUserRoles([]string{"content writer"}))
	cd.UserID = 1
	mock.ExpectQuery("SELECT 1\\s+FROM grants g\\s+JOIN roles r").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("(?s)WITH role_ids.*SELECT 1 FROM grants").WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow(1))
	CustomNewsIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("content writer with grant should see add news")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock.ExpectationsWereMet: %v", err)
	}
}

func TestCanPostNewsNilCoreData(t *testing.T) {
	if CanPostNews(nil) {
		t.Fatal("nil core data should not allow posting news")
	}
}
