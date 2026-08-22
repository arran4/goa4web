package forum

import (
	"context"
	"database/sql"
	"log"

	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

// UserCanCreateThread reports whether uid may create a thread in the topic.
func UserCanCreateThread(ctx context.Context, q db.Querier, section consts.PermissionSection, topicID, uid int32) (bool, error) {
	_, err := q.SystemCheckGrant(ctx, db.SystemCheckGrantParams{
		ViewerID:               uid,
		Section:                section.String(),
		Item:                   sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
		Action:                 consts.PermissionActionPost.String(),
		ItemID:                 sql.NullInt32{Int32: topicID, Valid: true},
		IsSpecificPrivateForum: (section.String() == "privateforum" || section.String() == "privateforum_thread") && topicID != 0,
		UserID:                 sql.NullInt32{Int32: uid, Valid: uid != 0},
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

func canForkThread(ctx context.Context, q db.Querier, section consts.PermissionSection, topicID, uid int32, sourceReplyable bool) (bool, error) {
	if !sourceReplyable {
		return false, nil
	}
	return UserCanCreateThread(ctx, q, section, topicID, uid)
}

func userCanReplyToThread(ctx context.Context, q db.Querier, section consts.PermissionSection, itemType consts.PermissionItem, itemID, threadID, uid int32) (bool, error) {
	thread, err := q.GetThreadBySectionThreadIDForReplier(ctx, db.GetThreadBySectionThreadIDForReplierParams{
		ReplierID:      uid,
		ThreadID:       threadID,
		Section:        section.String(),
		ItemType:       sql.NullString{String: itemType.String(), Valid: true},
		ItemID:         sql.NullInt32{Int32: itemID, Valid: itemID != 0},
		ReplierMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return thread != nil && (!thread.Locked.Valid || !thread.Locked.Bool), nil
}

// UserCanCreateTopic reports whether uid may create a topic in the category.
func UserCanCreateTopic(ctx context.Context, q db.Querier, section consts.PermissionSection, categoryID, uid int32) (bool, error) {
	_, err := q.SystemCheckGrant(ctx, db.SystemCheckGrantParams{
		ViewerID:               uid,
		Section:                section.String(),
		Item:                   sql.NullString{String: consts.PermissionItemCategory.String(), Valid: true},
		Action:                 consts.PermissionActionPost.String(),
		ItemID:                 sql.NullInt32{Int32: categoryID, Valid: true},
		IsSpecificPrivateForum: false,
		UserID:                 sql.NullInt32{Int32: uid, Valid: uid != 0},
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
func UserCanLabelTopic(ctx context.Context, q db.Querier, section consts.PermissionSection, topicID, uid int32) (bool, error) {
	_, err := q.SystemCheckGrant(ctx, db.SystemCheckGrantParams{
		ViewerID:               uid,
		Section:                section.String(),
		Item:                   sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
		Action:                 consts.PermissionActionLabel.String(),
		ItemID:                 sql.NullInt32{Int32: topicID, Valid: true},
		IsSpecificPrivateForum: (section.String() == "privateforum" || section.String() == "privateforum_thread") && topicID != 0,
		UserID:                 sql.NullInt32{Int32: uid, Valid: uid != 0},
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
