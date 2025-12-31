package common_test

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/core/common"
)

func TestHasGrant_AdminMode(t *testing.T) {
	// Setup a CoreData with an admin role but NO database queries (so explicit grants fail).
	// We use the "admin bypass" logic we are testing.

	// Need a dummy Querier to prevent nil pointer checks if they happen before the admin check (though typically they don't).
	// But in HasGrant:
	// if cd.HasAdminRole() { return true }
	// if cd.queries == nil { return false }
	// So nil querier is fine for testing the fallthrough.

	cd := common.NewCoreData(context.Background(), nil, nil, common.WithUserRoles([]string{"administrator"}))

	// Case 1: Admin Mode is ON
	cd.AdminMode = true
	if !cd.IsAdmin() {
		t.Fatal("Expected IsAdmin to be true")
	}
	if !cd.HasGrant("some_section", "some_item", "see", 123) {
		t.Error("Expected HasGrant to be true for admin in admin mode (bypass)")
	}

	// Case 2: Admin Mode is OFF
	cd.AdminMode = false
	if cd.IsAdmin() {
		t.Fatal("Expected IsAdmin to be false")
	}

	// This assertion should FAIL before the fix because HasGrant uses HasAdminRole() which is true.
	if cd.HasGrant("some_section", "some_item", "see", 123) {
		t.Error("Expected HasGrant to be false for admin NOT in admin mode (should not bypass)")
	}
}
