package writings

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/core/consts"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestAdminCategoryGrantsPage(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.AdminListRolesReturns = []*db.Role{
		{
			ID:            1,
			Name:          "user",
			CanLogin:      true,
			IsAdmin:       false,
			PrivateLabels: true,
		},
	}
	queries.ListGrantsReturns = []*db.Grant{
		{
			ID:       1,
			Section:  "writing",
			Item:     sql.NullString{String: "category", Valid: true},
			RuleType: "allow",
			ItemID:   sql.NullInt32{Int32: 1, Valid: true},
			Action:   "see",
			Active:   true,
		},
	}

	req := httptest.NewRequest("GET", "/admin/writings/categories/category/1/grants", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"category": "1"})
	rr := httptest.NewRecorder()

	AdminCategoryGrantsPage(rr, req)

	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}
