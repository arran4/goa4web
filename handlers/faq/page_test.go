package faq

import (
	"testing"

	corecommon "github.com/arran4/goa4web/core/common"
)

func TestCustomFAQIndexRoles(t *testing.T) {
	cd := corecommon.NewCoreData(nil, nil)
	cd.SetRoles([]string{"administrator"})
	cd.AdminMode = true
	CustomFAQIndex(cd)
	if !corecommon.ContainsItem(cd.CustomIndexItems, "Question Qontrols") {
		t.Errorf("admin should see question controls")
	}

	cd = corecommon.NewCoreData(nil, nil)
	cd.SetRoles([]string{"anonymous"})
	CustomFAQIndex(cd)
	if corecommon.ContainsItem(cd.CustomIndexItems, "Question Qontrols") {
		t.Errorf("anonymous should not see admin items")
	}
}
