package news

import (
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestCustomNewsIndexRoles(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		t.Run("Admin not in admin mode", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			cd := common.NewCoreData(req.Context(), testhelpers.NewQuerierStub(), config.NewRuntimeConfig(),
				common.WithUserRoles([]string{"administrator"}),
				common.WithPermissions([]*db.GetPermissionsByUserIDRow{
					{Name: "administrator", IsAdmin: true},
				}),
			)
			CustomNewsIndex(cd, req)
			if common.ContainsItem(cd.CustomIndexItems, "Add News") {
				t.Errorf("admin not in admin mode should not see add news")
			}
		})

		t.Run("Admin in admin mode", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			cd := common.NewCoreData(req.Context(), testhelpers.NewQuerierStub(), config.NewRuntimeConfig(),
				common.WithUserRoles([]string{"administrator"}),
				common.WithPermissions([]*db.GetPermissionsByUserIDRow{
					{Name: "administrator", IsAdmin: true},
				}),
			)
			cd.AdminMode = true
			CustomNewsIndex(cd, req)
			if !common.ContainsItem(cd.CustomIndexItems, "Add News") {
				t.Errorf("admin should see add news when admin mode is enabled")
			}
		})

		t.Run("Content writer with grant", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			ctx := req.Context()
			q := testhelpers.NewQuerierStub()
			q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
				if arg.Section == "news" && arg.Action == "post" && arg.Item.String == "post" {
					return 1, nil
				}
				return 0, sql.ErrNoRows
			}

			cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(),
				common.WithUserRoles([]string{"content writer"}),
			)
			cd.UserID = 1
			CustomNewsIndex(cd, req.WithContext(ctx))
			if !common.ContainsItem(cd.CustomIndexItems, "Add News") {
				t.Errorf("content writer with grant should see add news")
			}
		})

		t.Run("User without grant", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			q := testhelpers.NewQuerierStub()
			q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
				return 0, sql.ErrNoRows
			}

			cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
			CustomNewsIndex(cd, req)
			if common.ContainsItem(cd.CustomIndexItems, "Add News") {
				t.Errorf("user without grant should not see add news")
			}
		})
	})
}
