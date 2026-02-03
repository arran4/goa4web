package navigation

import (
	"testing"
	"github.com/arran4/goa4web/core/common"
)

func TestAdminSectionsHierarchy(t *testing.T) {
	r := NewRegistry()

	// Case 1: Nested sections
	r.RegisterAdminControlCenter("Core > Notifications", "Notify", "/notify", 10)
	// Case 2: Merging
	r.RegisterAdminControlCenter("Linker", "Linker", "/linker", 20)
	r.RegisterAdminControlCenter("Linker", "Other", "/other", 30)

	// Case 3: Deep nesting
	r.RegisterAdminControlCenter("Level1 > Level2 > Level3", "Item3", "/item3", 40)

	sections := r.AdminSections()

	// Helper to find section by name
	findSec := func(list []common.AdminSection, name string) *common.AdminSection {
		for i := range list {
			if list[i].Name == name {
				return &list[i]
			}
		}
		return nil
	}

	// 1. Core
	core := findSec(sections, "Core")
	if core == nil {
		t.Fatal("Core section not found")
	}
	if len(core.SubSections) != 1 || core.SubSections[0].Name != "Notifications" {
		t.Errorf("Expected Core > Notifications, got %#v", core.SubSections)
	}
	if len(core.SubSections[0].Links) != 1 || core.SubSections[0].Links[0].Name != "Notify" {
		t.Errorf("Expected Notify item, got %#v", core.SubSections[0].Links)
	}

	// 2. Linker
	linker := findSec(sections, "Linker")
	if linker == nil {
		t.Fatal("Linker section not found")
	}
	if linker.Link != "/linker" {
		t.Errorf("Linker section link mismatch. Want /linker, got %s", linker.Link)
	}
	// Check "Linker" item is removed
	for _, l := range linker.Links {
		if l.Name == "Linker" {
			t.Error("Linker item should be removed")
		}
	}
	// Check "Other" item exists
	foundOther := false
	for _, l := range linker.Links {
		if l.Name == "Other" {
			foundOther = true
			break
		}
	}
	if !foundOther {
		t.Error("Other item should exist")
	}

	// 3. Deep Nesting
	l1 := findSec(sections, "Level1")
	if l1 == nil {
		t.Fatal("Level1 not found")
	}
	if len(l1.SubSections) != 1 || l1.SubSections[0].Name != "Level2" {
		t.Fatal("Level2 not found")
	}
	l2 := &l1.SubSections[0]
	if len(l2.SubSections) != 1 || l2.SubSections[0].Name != "Level3" {
		t.Fatal("Level3 not found")
	}
	l3 := &l2.SubSections[0]
	if len(l3.Links) != 1 || l3.Links[0].Name != "Item3" {
		t.Fatal("Item3 not found")
	}
}
