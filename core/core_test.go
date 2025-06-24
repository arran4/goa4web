package core

import "testing"

func TestHasRole(t *testing.T) {
	cd := &CoreData{SecurityLevel: "moderator"}
	if !cd.HasRole("reader") {
		t.Errorf("moderator should have reader role")
	}
	if !cd.HasRole("moderator") {
		t.Errorf("moderator should have moderator role")
	}
	if cd.HasRole("administrator") {
		t.Errorf("moderator should not have administrator role")
	}
}
