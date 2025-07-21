package faq

import (
	"testing"

	"github.com/arran4/goa4web/core/common"
)

func TestCustomFAQIndexRoles(t *testing.T) {
	cd := common.NewCoreData(nil, nil)
	cd.SetRoles([]string{"administrator"})
	cd.AdminMode = true
	CustomFAQIndex(cd, nil)
	if !common.ContainsItem(cd.CustomIndexItems, "Question Qontrols") {
		t.Errorf("admin should see question controls")
	}

	cd = common.NewCoreData(nil, nil)
	cd.SetRoles([]string{"anonymous"})
	CustomFAQIndex(cd, nil)
	if common.ContainsItem(cd.CustomIndexItems, "Question Qontrols") {
		t.Errorf("anonymous should not see admin items")
	}
}
