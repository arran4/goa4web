package user

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

func setupRequest(t *testing.T, path string, userID int, queries db.Querier) (*http.Request, *common.CoreData) {
	t.Helper()
	req := httptest.NewRequest("GET", fmt.Sprintf(path, userID), nil)
	req = mux.SetURLVars(req, map[string]string{"user": strconv.Itoa(userID)})
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), queries, cfg, common.WithUserRoles([]string{"administrator"}))
	cd.LoadSelectionsFromRequest(req)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	return req, cd
}

func TestAdminUserPermissionsPage(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.SystemGetUserByIDFn = func(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
			if id != 2 {
				return nil, fmt.Errorf("unexpected user id: %d", id)
			}
			return &db.SystemGetUserByIDRow{Idusers: 2, Username: sql.NullString{String: "u", Valid: true}}, nil
		}
		queries.AdminListRolesReturns = []*db.Role{}
		queries.GetPermissionsByUserIDReturns = []*db.GetPermissionsByUserIDRow{}

		req, _ := setupRequest(t, "/admin/user/%d/permissions", 2, queries)
		rr := httptest.NewRecorder()
		adminUserPermissionsPage(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
	})
}

func TestAdminUserDisableConfirmPage(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.SystemGetUserByIDFn = func(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
			if id != 5 {
				return nil, fmt.Errorf("unexpected user id: %d", id)
			}
			return &db.SystemGetUserByIDRow{Idusers: 5, Username: sql.NullString{String: "u", Valid: true}}, nil
		}

		req, _ := setupRequest(t, "/admin/user/%d/disable", 5, queries)
		rr := httptest.NewRecorder()
		adminUserDisableConfirmPage(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
	})
}

func TestAdminUserEditFormPage(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.SystemGetUserByIDFn = func(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
			if id != 7 {
				return nil, fmt.Errorf("unexpected user id: %d", id)
			}
			return &db.SystemGetUserByIDRow{Idusers: 7, Username: sql.NullString{String: "u", Valid: true}}, nil
		}

		req, _ := setupRequest(t, "/admin/user/%d/edit", 7, queries)
		rr := httptest.NewRecorder()
		adminUserEditFormPage(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
	})
}
