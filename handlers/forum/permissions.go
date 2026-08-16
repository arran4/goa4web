package forum

import (
	"context"
	"database/sql"
	"log"

<<<<<<< HEAD
	"github.com/arran4/goa4web/core/consts"
=======
>>>>>>> 585b27a2 (feat(forum): implement post appending within time window)
	"github.com/arran4/goa4web/internal/db"
)

// UserCanCreateThread reports whether uid may create a thread in the topic.
<<<<<<< HEAD
func UserCanCreateThread(ctx context.Context, q db.Querier, section consts.PermissionSection, topicID, uid int32) (bool, error) {
	_, err := q.SystemCheckGrant(ctx, db.SystemCheckGrantParams{
		ViewerID: uid,
		Section:  section.String(),
		Item:     sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
		Action:   consts.PermissionActionPost.String(),
=======
func UserCanCreateThread(ctx context.Context, q db.Querier, section string, topicID, uid int32) (bool, error) {
	_, err := q.SystemCheckGrant(ctx, db.SystemCheckGrantParams{
		ViewerID: uid,
		Section:  section,
		Item:     sql.NullString{String: "topic", Valid: true},
		Action:   "post",
>>>>>>> 585b27a2 (feat(forum): implement post appending within time window)
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
<<<<<<< HEAD
func UserCanCreateTopic(ctx context.Context, q db.Querier, section consts.PermissionSection, categoryID, uid int32) (bool, error) {
	_, err := q.SystemCheckGrant(ctx, db.SystemCheckGrantParams{
		ViewerID: uid,
		Section:  section.String(),
		Item:     sql.NullString{String: consts.PermissionItemCategory.String(), Valid: true},
		Action:   consts.PermissionActionPost.String(),
=======
func UserCanCreateTopic(ctx context.Context, q db.Querier, section string, categoryID, uid int32) (bool, error) {
	_, err := q.SystemCheckGrant(ctx, db.SystemCheckGrantParams{
		ViewerID: uid,
		Section:  section,
		Item:     sql.NullString{String: "category", Valid: true},
		Action:   "post",
>>>>>>> 585b27a2 (feat(forum): implement post appending within time window)
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

// UserCanLabelTopic reports whether uid may add/remove labels on the topic.
<<<<<<< HEAD
func UserCanLabelTopic(ctx context.Context, q db.Querier, section consts.PermissionSection, topicID, uid int32) (bool, error) {
	_, err := q.SystemCheckGrant(ctx, db.SystemCheckGrantParams{
		ViewerID: uid,
		Section:  section.String(),
		Item:     sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
		Action:   consts.PermissionActionLabel.String(),
=======
func UserCanLabelTopic(ctx context.Context, q db.Querier, section string, topicID, uid int32) (bool, error) {
	_, err := q.SystemCheckGrant(ctx, db.SystemCheckGrantParams{
		ViewerID: uid,
		Section:  section,
		Item:     sql.NullString{String: "topic", Valid: true},
		Action:   "label",
>>>>>>> 585b27a2 (feat(forum): implement post appending within time window)
		ItemID:   sql.NullInt32{Int32: topicID, Valid: true},
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		// Log removed to avoid spam, or keep debug logging if needed
		return false, nil
	}
	log.Printf("UserCanLabelTopic error: uid=%d topic=%d err=%v", uid, topicID, err)
	return false, err
}
