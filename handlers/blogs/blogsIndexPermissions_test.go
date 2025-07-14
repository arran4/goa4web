package blogs

import (
	"net/http/httptest"
	"testing"

	corecommon "github.com/arran4/goa4web/core/common"
)

func TestCustomBlogIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs", nil)

	cd := corecommon.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"administrator"})
	cd.AdminMode = true
	CustomBlogIndex(cd, req)
	if !corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("admin should see user permissions")
	}
	if !corecommon.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("admin should see write blog")
	}

	cd = corecommon.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"content writer"})
	CustomBlogIndex(cd, req)
	if corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("content writer should not see user permissions")
	}
	if !corecommon.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("content writer should see write blog")
	}

	cd = corecommon.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"anonymous"})
	CustomBlogIndex(cd, req)
	if corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") || corecommon.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("anonymous should not see writer/admin items")
	}
}
