package blogs

import (
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
)

func TestCustomBlogIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs", nil)

	cd := common.NewCoreData(req.Context(), nil, common.WithConfig(config.NewRuntimeConfig()))
	cd.SetRoles([]string{"administrator"})
	cd.AdminMode = true
	CustomBlogIndex(cd, req)
	if !common.ContainsItem(cd.CustomIndexItems, "User Roles") {
		t.Errorf("admin should see user roles")
	}
	if !common.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("admin should see write blog")
	}

	cd = common.NewCoreData(req.Context(), nil, common.WithConfig(config.NewRuntimeConfig()))
	cd.SetRoles([]string{"content writer"})
	CustomBlogIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "User Roles") {
		t.Errorf("content writer should not see user roles")
	}
	if !common.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("content writer should see write blog")
	}

	cd = common.NewCoreData(req.Context(), nil, common.WithConfig(config.NewRuntimeConfig()))
	cd.SetRoles([]string{"anonymous"})
	CustomBlogIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "User Roles") || common.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("anonymous should not see writer/admin items")
	}
}
