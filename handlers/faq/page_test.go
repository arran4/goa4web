package faq

import (
	"testing"

	corecommon "github.com/arran4/goa4web/core/common"
)

func TestCustomFAQIndexRoles(t *testing.T) {
	cd := &corecommon.CoreData{SecurityLevel: "administrator", AdminMode: true}
	CustomFAQIndex(cd)
	if !corecommon.ContainsItem(cd.CustomIndexItems, "Question Qontrols") {
		t.Errorf("admin should see question controls")
	}

	cd = &corecommon.CoreData{SecurityLevel: "reader"}
	CustomFAQIndex(cd)
	if corecommon.ContainsItem(cd.CustomIndexItems, "Question Qontrols") {
		t.Errorf("reader should not see admin items")
	}
}
