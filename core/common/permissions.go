package common

import (
	"database/sql"

	"github.com/arran4/goa4web/internal/db"
)

const (
	// AdminAccessSection is the grants section controlling access to admin features.
	AdminAccessSection = "admin"
	// AdminAccessAction is the grants action required to reach admin routes.
	AdminAccessAction = "access"
)

// HasGrant reports whether the current user is allowed the given action.
func (cd *CoreData) HasGrant(section, item, action string, itemID int32) bool {
	if cd == nil {
		return false
	}
	if cd.IsAdmin() {
		return true
	}
	if cd.queries == nil {
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

// HasAdminAccess reports whether the caller can access admin functionality.
func (cd *CoreData) HasAdminAccess() bool {
	return cd.HasGrant(AdminAccessSection, "", AdminAccessAction, 0)
}
