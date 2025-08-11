package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminGrantsPageGroupsActions(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)

	grantsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "user_id", "role_id", "section", "item", "rule_type", "item_id", "item_rule", "action", "extra", "active"}).
		AddRow(1, nil, nil, 5, 7, "forum", "topic", "allow", 42, nil, "search", nil, true).
		AddRow(2, nil, nil, 5, 7, "forum", "topic", "allow", 42, nil, "view", nil, true)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, created_at, updated_at, user_id, role_id, section, item, rule_type, item_id, item_rule, action, extra, active FROM grants ORDER BY id")).WillReturnRows(grantsRows)

	userRows := sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).AddRow(5, nil, "bob", nil)
	mock.ExpectQuery("SELECT u\\.idusers").WithArgs(5).WillReturnRows(userRows)

	roleRows := sqlmock.NewRows([]string{"id", "name", "can_login", "is_admin", "private_labels", "public_profile_allowed_at"}).AddRow(7, "admin", true, false, true, nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, can_login, is_admin, private_labels, public_profile_allowed_at FROM roles WHERE id = ?")).WithArgs(7).WillReturnRows(roleRows)

	req := httptest.NewRequest("GET", "/admin/grants", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	AdminGrantsPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	body := rr.Body.String()
	if strings.Count(body, `<a href="/admin/user/5">bob (5)</a>`) != 1 {
		t.Fatalf("expected single user link: %s", body)
	}
	if !strings.Contains(body, `<a href="/admin/grant/1" class="pill">search</a>`) {
		t.Fatalf("missing search action: %s", body)
	}
	if !strings.Contains(body, `<a href="/admin/grant/2" class="pill">view</a>`) {
		t.Fatalf("missing view action: %s", body)
	}
}
