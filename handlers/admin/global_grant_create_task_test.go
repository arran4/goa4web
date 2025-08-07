package admin

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

// TestGlobalGrantCreateTask_ItemIDRequired verifies missing item_id errors.
func TestGlobalGrantCreateTask_ItemIDRequired(t *testing.T) {
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
	req := httptest.NewRequest("POST", "/admin/grant", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	cd := common.NewCoreData(context.Background(), db.New(conn), config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	if res := globalGrantCreateTask.Action(rr, req); res == nil {
		t.Fatalf("expected error, got nil")
	} else if err, ok := res.(error); !ok || err == nil {
		t.Fatalf("expected error, got %v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
