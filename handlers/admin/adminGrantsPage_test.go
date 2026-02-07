package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestHappyPathAdminGrantsPageGroupsActions(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.SearchGrantsReturns = []*db.SearchGrantsRow{
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
			Username: sql.NullString{String: "bob", Valid: true},
			RoleName: sql.NullString{String: "admin", Valid: true},
			CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
			UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
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
			Username: sql.NullString{String: "bob", Valid: true},
			RoleName: sql.NullString{String: "admin", Valid: true},
			CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
			UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		},
	}
	// Note: SystemGetUserByID and AdminGetRoleByID are not needed if SearchGrants returns the username/rolename.
	// But just in case the handler calls them for some reason (e.g. breadcrumbs or something, though unlikely for list page):
	// The original test stubbed them but only returned if ID matched, implying they might be called.
	// But wait, the original stub implementation of SearchGrants *used* them to populate the row.
	// If the handler calls SearchGrants, it gets the rows.
	// If the handler calls AdminGetRoleByID (e.g. if filtering by role), it might need it.
	// But here we are listing all grants.

	req := httptest.NewRequest("GET", "/admin/grants", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
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
