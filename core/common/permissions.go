package common

import (
	"database/sql"
	"github.com/arran4/goa4web/internal/db"
)

// HasGrant reports whether the current user is allowed the given action.
func (cd *CoreData) HasGrant(section, item, action string, itemID int32) bool {
	if cd.HasRole("administrator") {
		return true
	}
	if cd == nil || cd.queries == nil {
		return false
	}
	_, err := cd.queries.CheckGrant(cd.ctx, db.CheckGrantParams{
		ViewerID: cd.UserID,
		Section:  section,
		Item:     sql.NullString{String: item, Valid: item != ""},
		Action:   action,
		ItemID:   sql.NullInt32{Int32: itemID, Valid: itemID != 0},
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	return err == nil
}
