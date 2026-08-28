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

func TestGrantActionMap_EditAny(t *testing.T) {
	pubKey := "forum|topic"
	if def, ok := GrantActionMap[pubKey]; ok {
		hasEdit := false
		hasEditAny := false
		for _, a := range def.Actions {
			if a == consts.PermissionActionEdit.String() {
				hasEdit = true
			}
			if a == consts.PermissionActionEditAny.String() {
				hasEditAny = true
			}
		}
		if !hasEdit {
			t.Errorf("GrantActionMap[%q] is missing edit", pubKey)
		}
		if !hasEditAny {
			t.Errorf("GrantActionMap[%q] is missing edit-any", pubKey)
		}
	} else {
		t.Errorf("GrantActionMap does not contain %q", pubKey)
	}

	privKey := consts.PermissionSectionPrivateForumThread.String() + "|" + consts.PermissionItemThread.String()
	if def, ok := GrantActionMap[privKey]; ok {
		hasEdit := false
		hasEditAny := false
		for _, a := range def.Actions {
			if a == consts.PermissionActionEdit.String() {
				hasEdit = true
			}
			if a == consts.PermissionActionEditAny.String() {
				hasEditAny = true
			}
		}
		if !hasEdit {
			t.Errorf("GrantActionMap[%q] is missing edit", privKey)
		}
		if !hasEditAny {
			t.Errorf("GrantActionMap[%q] is missing edit-any", privKey)
		}
	} else {
		t.Errorf("GrantActionMap does not contain %q", privKey)
	}
}
