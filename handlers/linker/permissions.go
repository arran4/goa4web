package linker

import (
	"context"
	"database/sql"

	"github.com/arran4/goa4web/internal/db"
)

// UserCanCreateLink reports whether uid may submit a link to the category.
func UserCanCreateLink(ctx context.Context, q db.Querier, categoryID sql.NullInt32, uid int32) (bool, error) {
	_, err := q.SystemCheckGrant(ctx, db.SystemCheckGrantParams{
		ViewerID: uid,
		Section:  "linker",
		Item:     sql.NullString{String: "category", Valid: true},
		Action:   "post",
		ItemID:   categoryID,
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return false, err
}
