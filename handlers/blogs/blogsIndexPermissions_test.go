package blogs

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestCustomBlogIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs", nil)

	cd := common.NewCoreData(req.Context(), &db.QuerierStub{}, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}), common.WithPermissions([]*db.GetPermissionsByUserIDRow{
		{Name: "administrator", IsAdmin: true},
	}))
	cd.UserID = 1
	cd.AdminMode = true
	BlogsMiddlewareIndex(cd, req)
	if !common.ContainsItem(cd.CustomIndexItems, "Blogs Admin") {
		t.Errorf("admin should see blogs admin link")
	}
	if !common.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("admin should see write blog")
	}

	cd = common.NewCoreData(req.Context(), &db.QuerierStub{
		SystemCheckGrantReturns: 1,
	}, config.NewRuntimeConfig(), common.WithUserRoles([]string{"content writer"}))
	BlogsMiddlewareIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "Blogs Admin") {
		t.Errorf("content writer should not see blogs admin link")
	}
	if !common.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("content writer should see write blog")
	}

	cd = common.NewCoreData(req.Context(), &db.QuerierStub{SystemCheckGrantErr: errors.New("grant denied")}, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
	BlogsMiddlewareIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "Blogs Admin") || common.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("anyone should not see writer/admin items")
	}
}
