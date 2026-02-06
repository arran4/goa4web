package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestAdminAnyoneGrantsPage(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		qs := testhelpers.NewQuerierStub()
		qs.ListGrantsReturns = []*db.Grant{{
			ID:       1,
			Section:  "forum",
			Item:     sql.NullString{},
			RuleType: "allow",
			Action:   "search",
			Active:   true,
		}}

		req := httptest.NewRequest("GET", "/admin/grants/anyone", nil)
		ctx := req.Context()
		cfg := config.NewRuntimeConfig()
		cd := common.NewCoreData(ctx, qs, cfg, common.WithUserRoles([]string{"administrator"}), common.WithSilence(true))
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		AdminAnyoneGrantsPage(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		body := rr.Body.String()
		if !strings.Contains(body, `<a href="/admin/grants/anyone">Anyone</a>`) {
			t.Fatalf("missing link: %s", body)
		}
		if !strings.Contains(body, `<a href="/admin/grant/1" class="pill">search</a>`) {
			t.Fatalf("missing action pill: %s", body)
		}
	})
}
