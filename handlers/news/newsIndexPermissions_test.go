package news

import (
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
)

func TestCustomNewsIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	CustomNewsIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("admin not in admin mode should not see add news")
	}

	cd.AdminMode = true
	CustomNewsIndex(cd, req)
	if !common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("admin should see add news when admin mode is enabled")
	}

	type grantCheck struct {
		section string
		item    string
		action  string
		itemID  int32
	}
	var checks []grantCheck
	cd = common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(),
		common.WithUserRoles([]string{"content writer"}),
		common.WithGrantChecker(func(section, item, action string, itemID int32) bool {
			checks = append(checks, grantCheck{section: section, item: item, action: action, itemID: itemID})
			return true
		}))
	cd.UserID = 1
	CustomNewsIndex(cd, req)
	if !common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("content writer with grant should see add news")
	}
	if len(checks) != 1 || checks[0] != (grantCheck{section: "news", item: "post", action: "post", itemID: 0}) {
		t.Fatalf("grant checks = %+v, want one post grant check", checks)
	}

	checks = nil
	cd = common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(),
		common.WithUserRoles([]string{"anyone"}),
		common.WithGrantChecker(func(section, item, action string, itemID int32) bool {
			checks = append(checks, grantCheck{section: section, item: item, action: action, itemID: itemID})
			return false
		}))
	CustomNewsIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("user without grant should not see add news")
	}
	if len(checks) != 1 || checks[0] != (grantCheck{section: "news", item: "post", action: "post", itemID: 0}) {
		t.Fatalf("grant checks = %+v, want one post grant check", checks)
	}
}
