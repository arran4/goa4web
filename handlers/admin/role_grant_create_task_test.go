package admin

import (
	"context"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// TestRoleGrantCreateTask_ItemIDRequired verifies that grants needing an item ID
// fail when no item_id is supplied.
func TestRoleGrantCreateTask_ItemIDRequired(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	body := url.Values{
		"section": {"forum"},
		"item":    {"topic"},
		"action":  {"see"},
	}
	req := httptest.NewRequest("POST", "/admin/role/1/grant", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"role": "1"})

	cd := common.NewCoreData(context.Background(), db.New(conn), config.NewRuntimeConfig())
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	if res := roleGrantCreateTask.Action(rr, req); res == nil {
		t.Fatalf("expected error, got nil")
	} else if err, ok := res.(error); !ok || err == nil {
		t.Fatalf("expected error, got %v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

// TestRoleGrantCreateTask_MultipleActions verifies multiple action selections create
// a grant for each action.
func TestRoleGrantCreateTask_MultipleActions(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	body := url.Values{
		"section": {"forum"},
		"item":    {"topic"},
		"item_id": {"1"},
		"action":  {"see", "view"},
	}
	req := httptest.NewRequest("POST", "/admin/role/1/grant", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"role": "1"})

	insert := regexp.QuoteMeta("INSERT INTO grants (")
	mock.ExpectExec(insert).WithArgs(nil, int64(1), "forum", "topic", "allow", int64(1), nil, "see", nil).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(insert).WithArgs(nil, int64(1), "forum", "topic", "allow", int64(1), nil, "view", nil).WillReturnResult(sqlmock.NewResult(2, 1))

	cd := common.NewCoreData(context.Background(), db.New(conn), config.NewRuntimeConfig())
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	if res := roleGrantCreateTask.Action(rr, req); res == nil {
		t.Fatalf("expected response, got nil")
	} else if _, ok := res.(handlers.RefreshDirectHandler); !ok {
		t.Fatalf("expected RefreshDirectHandler, got %T", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
