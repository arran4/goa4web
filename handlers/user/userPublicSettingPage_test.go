package user

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
)

type publicProfileSettingsQueries struct {
	db.Querier
	user    *db.SystemGetUserByIDRow
	roleID  int32
	roleErr error
}

func (q *publicProfileSettingsQueries) SystemGetUserByID(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
	return q.user, nil
}

func (q *publicProfileSettingsQueries) GetPublicProfileRoleForUser(ctx context.Context, id int32) (int32, error) {
	if q.roleErr != nil {
		return 0, q.roleErr
	}
	return q.roleID, nil
}

func (q *publicProfileSettingsQueries) GetPreferenceForLister(ctx context.Context, id int32) (*db.Preference, error) {
	return nil, sql.ErrNoRows
}

func (q *publicProfileSettingsQueries) GetPermissionsByUserID(ctx context.Context, id int32) ([]*db.GetPermissionsByUserIDRow, error) {
	return nil, nil
}

func TestUserPublicProfileSettingPage_HasLink(t *testing.T) {
	queries := &publicProfileSettingsQueries{
		user: &db.SystemGetUserByIDRow{
			Idusers:  1,
			Username: sql.NullString{String: "testuser", Valid: true},
		},
		roleID: 1,
	}

	req := httptest.NewRequest("GET", "/usr/profile", nil)
	ctx := req.Context()
	cfg := config.NewRuntimeConfig()
	// Set template dir to relative path from this package
	cfg.TemplatesDir = "../../core/templates"

	cd := common.NewCoreData(ctx, queries, cfg, common.WithUserRoles([]string{}))
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	userPublicProfileSettingPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}

	body := rr.Body.String()
	expectedLink := "/user/profile/testuser"
	if !strings.Contains(body, expectedLink) {
		t.Errorf("Response body should contain link %q, but didn't. Body length: %d", expectedLink, len(body))
	}
}
