package faq

import (
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
)

func TestCustomFAQIndexRoles(t *testing.T) {
	cd := common.NewCoreData(nil, nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	cd.AdminMode = true
	CustomFAQIndex(cd, nil)
	if !common.ContainsItem(cd.CustomIndexItems, "Question Qontrols") {
		t.Errorf("admin should see question controls")
	}

	cd = common.NewCoreData(nil, nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anonymous"}))
	CustomFAQIndex(cd, nil)
	if common.ContainsItem(cd.CustomIndexItems, "Question Qontrols") {
		t.Errorf("anonymous should not see admin items")
	}
}
