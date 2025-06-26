package blogs

import (
	"net/http/httptest"
	"testing"

	corecommon "github.com/arran4/goa4web/core/common"
)

func TestCustomBlogIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs", nil)

	cd := &CoreData{SecurityLevel: "administrator"}
	CustomBlogIndex(cd, req)
	if !corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("admin should see user permissions")
	}
	if !corecommon.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("admin should see write blog")
	}

	cd = &CoreData{SecurityLevel: "writer"}
	CustomBlogIndex(cd, req)
	if corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("writer should not see user permissions")
	}
	if !corecommon.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("writer should see write blog")
	}

	cd = &CoreData{SecurityLevel: "reader"}
	CustomBlogIndex(cd, req)
	if corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") || corecommon.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("reader should not see writer/admin items")
	}
}
