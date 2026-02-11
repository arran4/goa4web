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
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestRoleGrantCreateTask(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		t.Run("Multiple Actions", func(t *testing.T) {
			q := testhelpers.NewQuerierStub()

			body := url.Values{
				"section": {"forum"},
				"item":    {"topic"},
				"item_id": {"1"},
				"action":  {"see", "view"},
			}
			req := httptest.NewRequest("POST", "/admin/role/1/grant", strings.NewReader(body.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req = mux.SetURLVars(req, map[string]string{"role": "1"})

			cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
			cd.LoadSelectionsFromRequest(req)
			ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			if res := roleGrantCreateTask.Action(rr, req); res == nil {
				t.Fatalf("expected response, got nil")
			} else if _, ok := res.(handlers.RefreshDirectHandler); !ok {
				t.Fatalf("expected RefreshDirectHandler, got %T", res)
			}
			if len(q.AdminCreateGrantCalls) != 2 {
				t.Fatalf("expected 2 grants, got %d", len(q.AdminCreateGrantCalls))
			}
			for _, grant := range q.AdminCreateGrantCalls {
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
		})
	})

	t.Run("Unhappy Path - Item ID Required", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		body := url.Values{
			"section": {"forum"},
			"item":    {"topic"},
			"action":  {"see"},
		}
		req := httptest.NewRequest("POST", "/admin/role/1/grant", strings.NewReader(body.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = mux.SetURLVars(req, map[string]string{"role": "1"})

		cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
		cd.LoadSelectionsFromRequest(req)
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		if res := roleGrantCreateTask.Action(rr, req); res == nil {
			t.Fatalf("expected error, got nil")
		} else if err, ok := res.(error); !ok || err == nil {
			t.Fatalf("expected error, got %v", res)
		}
	})
}
