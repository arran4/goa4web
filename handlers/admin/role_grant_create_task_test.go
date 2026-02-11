package admin

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

type roleGrantQueries struct {
	db.Querier
	created []db.AdminCreateGrantParams
}

func (q *roleGrantQueries) AdminCreateGrant(_ context.Context, arg db.AdminCreateGrantParams) (int64, error) {
	q.created = append(q.created, arg)
	return int64(len(q.created)), nil
}

// TestRoleGrantCreateTask_ItemIDRequired verifies that grants needing an item ID
// fail when no item_id is supplied.
func TestRoleGrantCreateTask_ItemIDRequired(t *testing.T) {
	body := url.Values{
		"section": {"forum"},
		"item":    {"topic"},
		"action":  {"see"},
	}
	req := httptest.NewRequest("POST", "/admin/role/1/grant", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"role": "1"})

	cd := common.NewCoreData(context.Background(), &roleGrantQueries{}, config.NewRuntimeConfig())
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	if res := roleGrantCreateTask.Action(rr, req); res == nil {
		t.Fatalf("expected error, got nil")
	} else if err, ok := res.(error); !ok || err == nil {
		t.Fatalf("expected error, got %v", res)
	}
}

// TestRoleGrantCreateTask_MultipleActions verifies multiple action selections create
// a grant for each action.
func TestRoleGrantCreateTask_MultipleActions(t *testing.T) {
	queries := &roleGrantQueries{}

	body := url.Values{
		"section": {"forum"},
		"item":    {"topic"},
		"item_id": {"1"},
		"action":  {"see", "view"},
	}
	req := httptest.NewRequest("POST", "/admin/role/1/grant", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"role": "1"})

	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	if res := roleGrantCreateTask.Action(rr, req); res == nil {
		t.Fatalf("expected response, got nil")
	} else if _, ok := res.(handlers.RefreshDirectHandler); !ok {
		t.Fatalf("expected RefreshDirectHandler, got %T", res)
	}
	if len(queries.created) != 2 {
		t.Fatalf("expected 2 grants, got %d", len(queries.created))
	}
	for _, grant := range queries.created {
		if !grant.RoleID.Valid || grant.RoleID.Int32 != 1 {
			t.Fatalf("unexpected role id: %#v", grant.RoleID)
		}
		if grant.Section != "forum" || grant.Item.String != "topic" || grant.Action == "" {
			t.Fatalf("unexpected grant: %#v", grant)
		}
		if grant.ItemID != (sql.NullInt32{Int32: 1, Valid: true}) {
			t.Fatalf("unexpected item id: %#v", grant.ItemID)
		}
	}
}
