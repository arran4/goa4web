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
)

type adminUserRouteQueries struct {
	db.Querier
	userID      int32
	user        *db.SystemGetUserByIDRow
	roles       []*db.Role
	permissions []*db.GetPermissionsByUserIDRow
}

func (q *adminUserRouteQueries) SystemGetUserByID(_ context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
	if id != q.userID {
		return nil, fmt.Errorf("unexpected user id: %d", id)
	}
	return q.user, nil
}

func (q *adminUserRouteQueries) AdminListRoles(context.Context) ([]*db.Role, error) {
	return q.roles, nil
}

func (q *adminUserRouteQueries) GetPermissionsByUserID(_ context.Context, id int32) ([]*db.GetPermissionsByUserIDRow, error) {
	if id != q.userID {
		return nil, fmt.Errorf("unexpected permissions user id: %d", id)
	}
	return q.permissions, nil
}

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

func TestAdminUserPermissionsPage_UserIDInjected(t *testing.T) {
	queries := &adminUserRouteQueries{
		userID:      2,
		user:        &db.SystemGetUserByIDRow{Idusers: 2, Username: sql.NullString{String: "u", Valid: true}},
		roles:       []*db.Role{},
		permissions: []*db.GetPermissionsByUserIDRow{},
	}
	req, _ := setupRequest(t, "/admin/user/%d/permissions", 2, queries)
	rr := httptest.NewRecorder()
	adminUserPermissionsPage(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestAdminUserDisableConfirmPage_UserIDInjected(t *testing.T) {
	queries := &adminUserRouteQueries{
		userID: 5,
		user:   &db.SystemGetUserByIDRow{Idusers: 5, Username: sql.NullString{String: "u", Valid: true}},
	}
	req, _ := setupRequest(t, "/admin/user/%d/disable", 5, queries)
	rr := httptest.NewRecorder()
	adminUserDisableConfirmPage(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestAdminUserEditFormPage_UserIDInjected(t *testing.T) {
	queries := &adminUserRouteQueries{
		userID: 7,
		user:   &db.SystemGetUserByIDRow{Idusers: 7, Username: sql.NullString{String: "u", Valid: true}},
	}
	req, _ := setupRequest(t, "/admin/user/%d/edit", 7, queries)
	rr := httptest.NewRecorder()
	adminUserEditFormPage(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}
