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
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestUserPublicProfileSettingPage_HasLink(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.SystemGetUserByIDFn = func(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
			return &db.SystemGetUserByIDRow{
				Idusers:  1,
				Username: sql.NullString{String: "testuser", Valid: true},
			}, nil
		}

		queries.GetPublicProfileRoleForUserFn = func(ctx context.Context, id int32) (int32, error) {
			return 1, nil
		}

		queries.GetPreferenceForListerFn = func(ctx context.Context, id int32) (*db.Preference, error) {
			return nil, sql.ErrNoRows
		}
		queries.GetPermissionsByUserIDReturns = []*db.GetPermissionsByUserIDRow{}

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
	})
}
