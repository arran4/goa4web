package news

import (
	"net/http/httptest"
	"testing"

	corecommon "github.com/arran4/goa4web/core/common"
)

func TestCustomNewsIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	cd := corecommon.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"administrator"})
	cd.AdminMode = true
	CustomNewsIndex(cd, req)
	if !corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("admin should see user permissions")
	}
	if !corecommon.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("admin should see add news")
	}

	cd = corecommon.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"content writer"})
	CustomNewsIndex(cd, req)
	if corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("content writer should not see user permissions")
	}
	if !corecommon.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("content writer should see add news")
	}

	cd = corecommon.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"anonymous"})
	CustomNewsIndex(cd, req)
	if corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") || corecommon.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("anonymous should not see admin items")
	}
}
