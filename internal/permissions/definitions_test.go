package permissions

import (
	"testing"

	"github.com/arran4/goa4web/core/consts"
)

func TestPermissionLookupAndGlobalValidity(t *testing.T) {
	tests := []struct {
		name          string
		section       string
		item          string
		action        string
		wantFound     bool
		wantRequireID bool
	}{
		{
			name:          "privateforum topic see is global",
			section:       "privateforum",
			item:          "topic",
			action:        "see",
			wantFound:     true,
			wantRequireID: false,
		},
		{
			name:          "privateforum topic create is global",
			section:       "privateforum",
			item:          "topic",
			action:        "create",
			wantFound:     true,
			wantRequireID: false,
		},
		{
			name:          "blogs entry post is global",
			section:       "blogs",
			item:          "entry",
			action:        "post",
			wantFound:     true,
			wantRequireID: false,
		},
		{
			name:          "forum topic post requires item ID",
			section:       "forum",
			item:          "topic",
			action:        "post",
			wantFound:     true,
			wantRequireID: true,
		},
		{
			name:          "forum thread edit requires item ID",
			section:       "forum",
			item:          "thread",
			action:        "edit",
			wantFound:     true,
			wantRequireID: true,
		},
		{
			name:          "privateforum_thread thread view requires item ID",
			section:       consts.PermissionSectionPrivateForumThread.String(),
			item:          consts.PermissionItemThread.String(),
			action:        consts.PermissionActionView.String(),
			wantFound:     true,
			wantRequireID: true,
		},
		{
			name:          "privateforum_thread thread reply requires item ID",
			section:       consts.PermissionSectionPrivateForumThread.String(),
			item:          consts.PermissionItemThread.String(),
			action:        consts.PermissionActionReply.String(),
			wantFound:     true,
			wantRequireID: true,
		},
		{
			name:          "unknown permission",
			section:       "unknown_section",
			item:          "unknown_item",
			action:        "unknown_action",
			wantFound:     false,
			wantRequireID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := Lookup(tt.section, tt.item, tt.action)
			if (def != nil) != tt.wantFound {
				t.Fatalf("Lookup(%q, %q, %q) found=%v, want %v", tt.section, tt.item, tt.action, def != nil, tt.wantFound)
			}
			if IsValid(tt.section, tt.item, tt.action) != tt.wantFound {
				t.Errorf("IsValid(%q, %q, %q) = %v, want %v", tt.section, tt.item, tt.action, IsValid(tt.section, tt.item, tt.action), tt.wantFound)
			}
			if def != nil {
				if def.RequireItemID != tt.wantRequireID {
					t.Errorf("RequireItemID = %v, want %v", def.RequireItemID, tt.wantRequireID)
				}
				wantGlobal := !tt.wantRequireID
				if IsValidGlobal(tt.section, tt.item, tt.action) != wantGlobal {
					t.Errorf("IsValidGlobal(%q, %q, %q) = %v, want %v", tt.section, tt.item, tt.action, IsValidGlobal(tt.section, tt.item, tt.action), wantGlobal)
				}
			} else {
				if IsValidGlobal(tt.section, tt.item, tt.action) {
					t.Errorf("IsValidGlobal(%q, %q, %q) expected false for unknown permission", tt.section, tt.item, tt.action)
				}
			}
		})
	}
}
