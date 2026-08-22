package common

import (
	"database/sql"

	"github.com/arran4/goa4web/internal/db"
)

// HasGrant reports whether the current user is allowed the given action.
func (cd *CoreData) HasGrant(section, item, action string, itemID int32) bool {
	if cd == nil {
		return false
	}
	if cd.IsAdmin() {
		if (section == "privateforum" || section == "privateforum_thread") && itemID != 0 {
			// Do not bypass permissions for specific private forum items, even for admins.
		} else {
			return true
		}
	}
	if cd.cache.testGrants != nil {
		for _, g := range cd.cache.testGrants {
			if g.Section == section && g.Action == action &&
				(!g.Item.Valid || g.Item.String == item || g.Item.String == "") &&
				g.Active {

				if (section == "privateforum" || section == "privateforum_thread") && itemID != 0 {
					if g.ItemID.Valid && g.ItemID.Int32 == itemID && g.UserID.Valid && g.UserID.Int32 == cd.UserID {
						return true
					}
				} else if !g.ItemID.Valid || g.ItemID.Int32 == itemID || g.ItemID.Int32 == 0 {
					return true
				}
			}
		}
	}

	if cd.queries == nil {
		return false
	}

	_, err := cd.queries.SystemCheckGrant(cd.ctx, db.SystemCheckGrantParams{
		ViewerID:               cd.UserID,
		Section:                section,
		Item:                   sql.NullString{String: item, Valid: item != ""},
		Action:                 action,
		ItemID:                 sql.NullInt32{Int32: itemID, Valid: itemID != 0},
		IsSpecificPrivateForum: (section == "privateforum" || section == "privateforum_thread") && itemID != 0,
		UserID:                 sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	return err == nil
}
