package permissions

import (
	"testing"

	"github.com/arran4/goa4web/core/consts"
)

func TestDefinitionsIncludePrivateForumThreadPermissions(t *testing.T) {
	want := map[consts.PermissionAction]bool{
		consts.PermissionActionView:  false,
		consts.PermissionActionReply: false,
	}
	for _, definition := range Definitions {
		if definition.Section != consts.PermissionSectionPrivateForumThread.String() || definition.Item != consts.PermissionItemThread.String() {
			continue
		}
		action := consts.PermissionAction(definition.Action)
		if _, ok := want[action]; ok {
			want[action] = true
		}
	}
	for action, found := range want {
		if !found {
			t.Errorf("permission definitions missing private thread %q", action)
		}
	}
}
