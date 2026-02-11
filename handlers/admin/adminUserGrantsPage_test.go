package admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestHappyPathAdminUserGrantsPage_UserIDInjected(t *testing.T) {
	userID := 3
	queries := testhelpers.NewQuerierStub()
	queries.SystemGetUserByIDFn = func(_ context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
		if id != int32(userID) {
			return nil, fmt.Errorf("unexpected user id: %d", id)
		}
		return &db.SystemGetUserByIDRow{
			Idusers:                int32(userID),
			Username:               sql.NullString{String: "u", Valid: true},
			PublicProfileEnabledAt: sql.NullTime{},
		}, nil
	}
	queries.GetPermissionsByUserIDReturns = []*db.GetPermissionsByUserIDRow{}
	queries.ListGrantsByUserIDReturns = []*db.Grant{}
	queries.GetAllForumCategoriesReturns = []*db.Forumcategory{}
	queries.SystemListLanguagesReturns = []*db.Language{}

	req := httptest.NewRequest("GET", fmt.Sprintf("/admin/user/%d/grants", userID), nil)
	req = mux.SetURLVars(req, map[string]string{"user": strconv.Itoa(userID)})
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), queries, cfg)
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	(&AdminUserGrantsPage{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}
