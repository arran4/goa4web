package forum

import (
	"context"
	"database/sql"
	"log"

	"github.com/arran4/goa4web/internal/db"
)

// UserCanCreateThread reports whether uid may create a thread in the topic.
func UserCanCreateThread(ctx context.Context, q db.Querier, topicID, uid int32) (bool, error) {
	_, err := q.SystemCheckGrant(ctx, db.SystemCheckGrantParams{
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
		log.Printf("UserCanCreateThread deny: uid=%d topic=%d", uid, topicID)
		return false, nil
	}
	log.Printf("UserCanCreateThread error: uid=%d topic=%d err=%v", uid, topicID, err)
	return false, err
}

// UserCanCreateTopic reports whether uid may create a topic in the category.
func UserCanCreateTopic(ctx context.Context, q db.Querier, categoryID, uid int32) (bool, error) {
	_, err := q.SystemCheckGrant(ctx, db.SystemCheckGrantParams{
		ViewerID: uid,
		Section:  "forum",
		Item:     sql.NullString{String: "category", Valid: true},
		Action:   "post",
		ItemID:   sql.NullInt32{Int32: categoryID, Valid: true},
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		log.Printf("UserCanCreateTopic deny: uid=%d category=%d", uid, categoryID)
		return false, nil
	}
	log.Printf("UserCanCreateTopic error: uid=%d category=%d err=%v", uid, categoryID, err)
	return false, err
}
