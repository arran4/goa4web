package blogs

import (
	"net/http/httptest"
	"testing"

	common "github.com/arran4/goa4web/core/common"
)

func TestCustomBlogIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs", nil)

	cd := common.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"administrator"})
	cd.AdminMode = true
	CustomBlogIndex(cd, req)
	if !common.ContainsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("admin should see user permissions")
	}
	if !common.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("admin should see write blog")
	}

	cd = common.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"content writer"})
	CustomBlogIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("content writer should not see user permissions")
	}
	if !common.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("content writer should see write blog")
	}

	cd = common.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"anonymous"})
	CustomBlogIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "User Permissions") || common.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("anonymous should not see writer/admin items")
	}
}
