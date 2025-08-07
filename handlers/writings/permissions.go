package writings

import (
	"context"
	"database/sql"

	"github.com/arran4/goa4web/internal/db"
)

// UserCanCreateWriting reports whether uid may publish an article in the category.
func UserCanCreateWriting(ctx context.Context, q db.Querier, categoryID, uid int32) (bool, error) {
	_, err := q.SystemCheckGrant(ctx, db.SystemCheckGrantParams{
		ViewerID: uid,
		Section:  "writing",
		Item:     sql.NullString{String: "category", Valid: true},
		Action:   "post",
		ItemID:   sql.NullInt32{Int32: categoryID, Valid: true},
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
