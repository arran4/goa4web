package news

import (
	"net/http/httptest"
	"testing"

	corecommon "github.com/arran4/goa4web/core/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
)

func TestCustomNewsIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	cd := &hcommon.CoreData{SecurityLevel: "administrator", AdminMode: true}
	CustomNewsIndex(cd, req)
	if !corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("admin should see user permissions")
	}
	if !corecommon.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("admin should see add news")
	}

	cd = &hcommon.CoreData{SecurityLevel: "writer"}
	CustomNewsIndex(cd, req)
	if corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("writer should not see user permissions")
	}
	if !corecommon.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("writer should see add news")
	}

	cd = &hcommon.CoreData{SecurityLevel: "reader"}
	CustomNewsIndex(cd, req)
	if corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") || corecommon.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("reader should not see admin items")
	}
}
