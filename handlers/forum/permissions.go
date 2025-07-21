package forum

import (
	"context"
	"database/sql"

	"github.com/arran4/goa4web/internal/db"
)

// UserCanCreateThread reports whether uid may create a thread in the topic.
func UserCanCreateThread(ctx context.Context, q *db.Queries, topicID, uid int32) (bool, error) {
	_, err := q.CheckGrant(ctx, db.CheckGrantParams{
		ViewerID: uid,
		Section:  "forum",
		Item:     sql.NullString{String: "topic", Valid: true},
		Action:   "post",
		ItemID:   sql.NullInt32{Int32: topicID, Valid: true},
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
