package admin

import (
	"slices"
	"testing"

	"github.com/arran4/goa4web/core/consts"
)

func TestGrantActionMapIncludesForumMutationActions(t *testing.T) {
	keys := []string{
		consts.PermissionSectionForum.String() + "|" + consts.PermissionItemTopic.String(),
		consts.PermissionSectionPrivateForumThread.String() + "|" + consts.PermissionItemThread.String(),
	}
	wantActions := []string{
		consts.PermissionActionAppend.String(),
		consts.PermissionActionEdit.String(),
		consts.PermissionActionEditAny.String(),
	}
	for _, key := range keys {
		definition, ok := GrantActionMap[key]
		if !ok {
			t.Fatalf("GrantActionMap missing %q", key)
		}
		for _, action := range wantActions {
			if !slices.Contains(definition.Actions, action) {
				t.Errorf("GrantActionMap[%q] missing %q", key, action)
			}
		}
	}
}
