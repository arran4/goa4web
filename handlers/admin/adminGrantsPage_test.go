package admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

type grantsPageQueries struct {
	db.Querier
	grants []*db.Grant
	userID int32
	user   *db.SystemGetUserByIDRow
	roleID int32
	role   *db.Role
}

func (q *grantsPageQueries) ListGrants(context.Context) ([]*db.Grant, error) {
	return q.grants, nil
}

func (q *grantsPageQueries) SystemCheckGrant(_ context.Context, arg db.SystemCheckGrantParams) (int32, error) {
	if arg.Section == common.AdminAccessSection && arg.Action == common.AdminAccessAction {
		return 1, nil
	}
	return 0, fmt.Errorf("unexpected grant check: %#v", arg)
}

func (q *grantsPageQueries) SystemGetUserByID(_ context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
	if id != q.userID {
		return nil, fmt.Errorf("unexpected user id: %d", id)
	}
	return q.user, nil
}

func (q *grantsPageQueries) AdminGetRoleByID(_ context.Context, id int32) (*db.Role, error) {
	if id != q.roleID {
		return nil, fmt.Errorf("unexpected role id: %d", id)
	}
	return q.role, nil
}

func TestAdminGrantsPageGroupsActions(t *testing.T) {
	queries := &grantsPageQueries{
		userID: 5,
		roleID: 7,
		grants: []*db.Grant{
			{
				ID:       1,
				UserID:   sql.NullInt32{Int32: 5, Valid: true},
				RoleID:   sql.NullInt32{Int32: 7, Valid: true},
				Section:  "forum",
				Item:     sql.NullString{String: "topic", Valid: true},
				RuleType: "allow",
				ItemID:   sql.NullInt32{Int32: 42, Valid: true},
				Action:   "search",
				Active:   true,
			},
			{
				ID:       2,
				UserID:   sql.NullInt32{Int32: 5, Valid: true},
				RoleID:   sql.NullInt32{Int32: 7, Valid: true},
				Section:  "forum",
				Item:     sql.NullString{String: "topic", Valid: true},
				RuleType: "allow",
				ItemID:   sql.NullInt32{Int32: 42, Valid: true},
				Action:   "view",
				Active:   true,
			},
		},
		user: &db.SystemGetUserByIDRow{
			Idusers:                5,
			Username:               sql.NullString{String: "bob", Valid: true},
			PublicProfileEnabledAt: sql.NullTime{},
		},
		role: &db.Role{Name: "admin", CanLogin: true, IsAdmin: false, PrivateLabels: true},
	}

	req := httptest.NewRequest("GET", "/admin/grants", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{}))
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	AdminGrantsPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()
	if strings.Count(body, `<a href="/admin/user/5">bob (5)</a>`) != 1 {
		t.Fatalf("expected single user link: %s", body)
	}
	if !strings.Contains(body, `<a href="/admin/grant/1" class="pill">search</a>`) {
		t.Fatalf("missing search action: %s", body)
	}
	if !strings.Contains(body, `<a href="/admin/grant/2" class="pill">view</a>`) {
		t.Fatalf("missing view action: %s", body)
	}
}
