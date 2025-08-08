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

func TestAdminAnyoneGrantsPage(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)

	grantsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "user_id", "role_id", "section", "item", "rule_type", "item_id", "item_rule", "action", "extra", "active"}).
		AddRow(1, nil, nil, nil, nil, "forum", "", "allow", nil, nil, "search", nil, true)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, created_at, updated_at, user_id, role_id, section, item, rule_type, item_id, item_rule, action, extra, active FROM grants ORDER BY id")).WillReturnRows(grantsRows)

	req := httptest.NewRequest("GET", "/admin/grants/anyone", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	AdminAnyoneGrantsPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `<a href="/admin/grants/anyone">Anyone</a>`) {
		t.Fatalf("missing link: %s", body)
	}
}
