package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// JobTypeForumReplyNotification identifies committed forum reply activities awaiting notification delivery.
const JobTypeForumReplyNotification = "forum_reply_notification"

// RecordForumReplyNotification stores a reply's identifiers using the caller's transaction.
// The caller signals the job worker after committing successfully.
func RecordForumReplyNotification(ctx context.Context, q db.Querier, activity string, position int64, commentID, actorID int32) error {
	dedupeKey := DedupeKey(JobTypeForumReplyNotification, activity)
	_, err := q.SystemInsertDurableJob(ctx, db.SystemInsertDurableJobParams{
		JobType:         JobTypeForumReplyNotification,
		DedupeKey:       dedupeKey,
		ActiveDedupeKey: sql.NullString{String: dedupeKey, Valid: true},
		TargetType:      "comment",
		TargetID:        int64(commentID),
		TargetKey:       activity,
		TargetRevision:  position,
		MaxAttempts:     DefaultMaxAttempts,
		RunAfter:        time.Now().UTC(),
		CreatedByUserID: sql.NullInt32{Int32: actorID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("record forum reply notification: %w", err)
	}
	return nil
}
