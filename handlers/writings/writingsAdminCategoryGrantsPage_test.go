package writings

import (
	"context"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gorilla/mux"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

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
	mock.ExpectQuery("SELECT id, name, can_login, is_admin, public_profile_allowed_at FROM roles ORDER BY id").WillReturnRows(rolesRows)

	grantsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "user_id", "role_id", "section", "item", "rule_type", "item_id", "item_rule", "action", "extra", "active"}).
		AddRow(1, nil, nil, nil, nil, "writing", "category", "allow", 1, nil, "see", nil, true)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, created_at, updated_at, user_id, role_id, section, item, rule_type, item_id, item_rule, action, extra, active FROM grants ORDER BY id")).WillReturnRows(grantsRows)

	req := httptest.NewRequest("GET", "/admin/writings/categories/category/1/grants", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"category": "1"})
	rr := httptest.NewRecorder()

	AdminCategoryGrantsPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}
