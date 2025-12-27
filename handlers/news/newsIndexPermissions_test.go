package news

import (
	"database/sql"
	"net/http/httptest"
	"regexp"
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

	conn, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	cd = common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
	mock.ExpectQuery(regexp.QuoteMeta("WITH role_ids AS (\n    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = ?\n    UNION\n    SELECT id FROM roles WHERE name = 'anyone'\n)\nSELECT 1 FROM grants g\nWHERE g.section = ?\n  AND (g.item = ? OR g.item IS NULL)\n  AND g.action = ?\n  AND g.active = 1\n  AND (g.item_id = ? OR g.item_id IS NULL)\n  AND (g.user_id = ? OR g.user_id IS NULL)\n  AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))\nLIMIT 1\n")).
		WithArgs(int32(0), "news", sql.NullString{String: "post", Valid: true}, "post", sql.NullInt32{}, sql.NullInt32{}).
		WillReturnError(sql.ErrNoRows)
	CustomNewsIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("user without grant should not see add news")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("grant expectations: %v", err)
	}
}
