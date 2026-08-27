package admin

import (
	"github.com/arran4/goa4web/core/consts"
	"testing"
)

func TestGrantActionMap_AppendPresent(t *testing.T) {
	// Test that public forum topics include append
	if def, ok := GrantActionMap["forum|topic"]; ok {
		hasAppend := false
		for _, action := range def.Actions {
			if action == consts.PermissionActionAppend.String() {
				hasAppend = true
				break
			}
		}
		if !hasAppend {
			t.Errorf("GrantActionMap[\"forum|topic\"] is missing the append action")
		}
	} else {
		t.Errorf("GrantActionMap does not contain \"forum|topic\"")
	}

	// Test that private forum threads include append
	privateKey := consts.PermissionSectionPrivateForumThread.String() + "|" + consts.PermissionItemThread.String()
	if def, ok := GrantActionMap[privateKey]; ok {
		hasAppend := false
		for _, action := range def.Actions {
			if action == consts.PermissionActionAppend.String() {
				hasAppend = true
				break
			}
		}
		if !hasAppend {
			t.Errorf("GrantActionMap[%q] is missing the append action", privateKey)
		}
	} else {
		t.Errorf("GrantActionMap does not contain %q", privateKey)
	}
}
