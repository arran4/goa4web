package admin

import (
	"slices"
	"testing"

	"github.com/arran4/goa4web/core/consts"
)

func TestGrantActionMapIncludesPrivateForumThreads(t *testing.T) {
	key := consts.PermissionSectionPrivateForumThread.String() + "|" + consts.PermissionItemThread.String()
	definition, ok := GrantActionMap[key]
	if !ok {
		t.Fatalf("GrantActionMap missing %q", key)
	}
	if !definition.RequireItemID {
		t.Error("private thread grants must require an item ID")
	}
	for _, action := range []string{consts.PermissionActionView.String(), consts.PermissionActionReply.String()} {
		if !slices.Contains(definition.Actions, action) {
			t.Errorf("private thread grants missing %q", action)
		}
	}
}
