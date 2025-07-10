package faq

import (
	"testing"

	corecommon "github.com/arran4/goa4web/core/common"
)

func TestCustomFAQIndexRoles(t *testing.T) {
	cd := corecommon.NewCoreData(nil, nil)
	cd.SetRole("administrator")
	cd.AdminMode = true
	CustomFAQIndex(cd)
	if !corecommon.ContainsItem(cd.CustomIndexItems, "Question Qontrols") {
		t.Errorf("admin should see question controls")
	}

	cd = corecommon.NewCoreData(nil, nil)
	cd.SetRole("reader")
	CustomFAQIndex(cd)
	if corecommon.ContainsItem(cd.CustomIndexItems, "Question Qontrols") {
		t.Errorf("reader should not see admin items")
	}
}
