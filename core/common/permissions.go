package common

import (
	"database/sql"

	"github.com/arran4/goa4web/internal/db"
)

const (
	// AdminGrantSection identifies the grant section used for administrator access.
	AdminGrantSection = "admin"
	// AdminGrantAccessAction is the grant action that unlocks administrator pages.
	AdminGrantAccessAction = "access"
)

// HasGrant reports whether the current user is allowed the given action.
func (cd *CoreData) HasGrant(section, item, action string, itemID int32) bool {
	if cd == nil {
		return false
	}
	if cd.IsAdmin() {
		return true
	}
	return cd.checkGrant(section, item, action, itemID)
}

// HasAdminAccess reports whether the caller can reach administrator-only pages.
func (cd *CoreData) HasAdminAccess() bool {
	if cd == nil {
		return false
	}
	if cd.HasAdminRole() {
		return true
	}
	if cd.AdminMode && cd.IsAdmin() {
		return true
	}
	return cd.checkGrant(AdminGrantSection, "", AdminGrantAccessAction, 0) ||
		cd.checkGrant("role", "", "admin", 0)
}

func (cd *CoreData) checkGrant(section, item, action string, itemID int32) bool {
	if cd == nil || cd.queries == nil {
		return false
	}
	_, err := cd.queries.SystemCheckGrant(cd.ctx, db.SystemCheckGrantParams{
		ViewerID: cd.UserID,
		Section:  section,
		Item:     sql.NullString{String: item, Valid: item != ""},
		Action:   action,
		ItemID:   sql.NullInt32{Int32: itemID, Valid: itemID != 0},
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	return err == nil
}
