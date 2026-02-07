package blogs

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestHappyPathCustomBlogIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs", nil)

	t.Run("Administrator", func(t *testing.T) {
		cd := common.NewCoreData(req.Context(), testhelpers.NewQuerierStub(
			testhelpers.FromScenario(testhelpers.ScenarioAdmin()),
		), config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
		cd.UserID = 1
		cd.AdminMode = true
		BlogsMiddlewareIndex(cd, req)
		if !common.ContainsItem(cd.CustomIndexItems, "Blogs Admin") {
			t.Errorf("admin should see blogs admin link")
		}
		if !common.ContainsItem(cd.CustomIndexItems, "Write blog") {
			t.Errorf("admin should see write blog")
		}
	})

	t.Run("Content Writer", func(t *testing.T) {
		cd := common.NewCoreData(req.Context(), testhelpers.NewQuerierStub(
			testhelpers.WithGrantResult(true),
		), config.NewRuntimeConfig(), common.WithUserRoles([]string{"content writer"}))
		BlogsMiddlewareIndex(cd, req)
		if common.ContainsItem(cd.CustomIndexItems, "Blogs Admin") {
			t.Errorf("content writer should not see blogs admin link")
		}
		if !common.ContainsItem(cd.CustomIndexItems, "Write blog") {
			t.Errorf("content writer should see write blog")
		}
	})

	t.Run("Anyone", func(t *testing.T) {
		cd := common.NewCoreData(req.Context(), testhelpers.NewQuerierStub(
			testhelpers.WithGrantError(errors.New("grant denied")),
		), config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
		BlogsMiddlewareIndex(cd, req)
		if common.ContainsItem(cd.CustomIndexItems, "Blogs Admin") || common.ContainsItem(cd.CustomIndexItems, "Write blog") {
			t.Errorf("anyone should not see writer/admin items")
		}
	})
}
