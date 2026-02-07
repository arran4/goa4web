package faq

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
)

func TestHappyPathCustomFAQIndexRoles(t *testing.T) {
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	cd.AdminMode = true
	CustomFAQIndex(cd, nil)
	if common.ContainsItem(cd.CustomIndexItems, "Question Qontrols") {
		t.Errorf("admin should not see question controls")
	}

	cd = common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
	CustomFAQIndex(cd, nil)
	if common.ContainsItem(cd.CustomIndexItems, "Question Qontrols") {
		t.Errorf("anyone should not see admin items")
	}
}
