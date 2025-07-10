package news

import (
	"net/http/httptest"
	"testing"

	corecommon "github.com/arran4/goa4web/core/common"
)

func TestCustomNewsIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	cd := corecommon.NewCoreData(req.Context(), nil)
	cd.SetRole("administrator")
	cd.AdminMode = true
	CustomNewsIndex(cd, req)
	if !corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("admin should see user permissions")
	}
	if !corecommon.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("admin should see add news")
	}

	cd = corecommon.NewCoreData(req.Context(), nil)
	cd.SetRole("writer")
	CustomNewsIndex(cd, req)
	if corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") {
		t.Errorf("writer should not see user permissions")
	}
	if !corecommon.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("writer should see add news")
	}

	cd = corecommon.NewCoreData(req.Context(), nil)
	cd.SetRole("reader")
	CustomNewsIndex(cd, req)
	if corecommon.ContainsItem(cd.CustomIndexItems, "User Permissions") || corecommon.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("reader should not see admin items")
	}
}
