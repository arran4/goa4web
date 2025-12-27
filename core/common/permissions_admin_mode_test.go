package common_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

type adminStub struct {
	db.Querier
}

func (a *adminStub) GetAdministratorUserRole(ctx context.Context, usersIdusers int32) (*db.UserRole, error) {
	return &db.UserRole{}, nil
}

func (a *adminStub) SystemCheckRoleGrant(ctx context.Context, arg db.SystemCheckRoleGrantParams) (int32, error) {
	return 0, sql.ErrNoRows
}

func (a *adminStub) SystemCheckGrant(ctx context.Context, arg db.SystemCheckGrantParams) (int32, error) {
	return 0, sql.ErrNoRows
}

func TestHasGrant_AdminMode(t *testing.T) {
	// Setup a CoreData with an admin role but NO database queries (so explicit grants fail).
	// We use the "admin bypass" logic we are testing.

	// Need a dummy Querier to prevent nil pointer checks if they happen before the admin check (though typically they don't).
	// But in HasGrant:
	// if cd.HasAdminRole() { return true }
	// if cd.queries == nil { return false }
	// So nil querier is fine for testing the fallthrough.

	cd := common.NewCoreData(context.Background(), &adminStub{}, nil, common.WithUserRoles([]string{"administrator"}))
	cd.UserID = 1

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
