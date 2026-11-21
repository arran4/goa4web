package common

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/internal/db"
)

// CalculateActivityPosition returns the monotonically ordered 64-bit position
// for a thread activity.
func CalculateActivityPosition(commentID int32, appendID int64) int64 {
	if commentID <= 0 {
		return 0
	}
	if appendID < 0 {
		appendID = 0
	}
	return (int64(commentID) << 32) | (appendID & 0xFFFFFFFF)
}

// ActivityPositionCommentAndAppend decomposes a 64-bit activity position into
// its comment ID and append ID.
func ActivityPositionCommentAndAppend(pos int64) (commentID int32, appendID int64) {
	if pos <= 0 {
		return 0, 0
	}
	return int32(pos >> 32), pos & 0xFFFFFFFF
}

// ParseActivityPosition parses a canonical activity ID (e.g. "comment:10",
// "comment:10:append:2", "reply:comment:10:append:2") or numeric position string.
func ParseActivityPosition(activity string) (int64, error) {
	activity = strings.TrimSpace(activity)
	if activity == "" {
		return 0, nil
	}
	activity = strings.TrimPrefix(activity, "reply:")
	if strings.HasPrefix(activity, "comment:") {
		rem := strings.TrimPrefix(activity, "comment:")
		if idx := strings.Index(rem, ":append:"); idx != -1 {
			cidStr := rem[:idx]
			aidStr := rem[idx+len(":append:"):]
			cid, err := strconv.ParseInt(cidStr, 10, 32)
			if err != nil {
				return 0, fmt.Errorf("invalid comment ID in activity %q: %w", activity, err)
			}
			aid, err := strconv.ParseInt(aidStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid append ID in activity %q: %w", activity, err)
			}
			return CalculateActivityPosition(int32(cid), aid), nil
		}
		if strings.HasSuffix(rem, ":append") {
			cidStr := strings.TrimSuffix(rem, ":append")
			cid, err := strconv.ParseInt(cidStr, 10, 32)
			if err != nil {
				return 0, fmt.Errorf("invalid comment ID in activity %q: %w", activity, err)
			}
			return CalculateActivityPosition(int32(cid), 0), nil
		}
		cid, err := strconv.ParseInt(rem, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("invalid comment ID in activity %q: %w", activity, err)
		}
		return CalculateActivityPosition(int32(cid), 0), nil
	}
	val, err := strconv.ParseInt(activity, 10, 64)
	if err == nil {
		if val > 0xFFFFFFFF {
			return val, nil
		}
		return CalculateActivityPosition(int32(val), 0), nil
	}
	return 0, fmt.Errorf("cannot parse activity position %q", activity)
}

// FormatActivityPosition formats an activity position into its canonical string representation.
func FormatActivityPosition(commentID int32, appendID int64) string {
	if commentID <= 0 {
		return ""
	}
	if appendID > 0 {
		return fmt.Sprintf("comment:%d:append:%d", commentID, appendID)
	}
	return fmt.Sprintf("comment:%d", commentID)
}

// CommentLastActivity returns the canonical activity ID for the latest state of commentID.
// If the comment was loaded as part of thread comments in the current context, it returns the
// snapshot-consistent activity identity that was rendered to the user.
func (cd *CoreData) CommentLastActivity(commentID int32) string {
	if cd == nil || commentID <= 0 {
		return fmt.Sprintf("comment:%d", commentID)
	}
	if cd.cache.forumThreadComments != nil {
		cd.cache.mapMu.Lock()
		for _, lv := range cd.cache.forumThreadComments {
			if lv != nil {
				if comments, ok := lv.Peek(); ok {
					for _, c := range comments {
						if c.Idcomments == commentID {
							cd.cache.mapMu.Unlock()
							return FormatActivityPosition(c.Idcomments, c.LastAppendID)
						}
					}
				}
			}
		}
		cd.cache.mapMu.Unlock()
	}
	if cd.queries == nil {
		return fmt.Sprintf("comment:%d", commentID)
	}
	appends, err := cd.queries.SystemListCommentAppendsByCommentID(cd.ctx, commentID)
	if err == nil && len(appends) > 0 {
		lastAppend := appends[len(appends)-1]
		return fmt.Sprintf("comment:%d:append:%d", commentID, lastAppend.ID)
	}
	return fmt.Sprintf("comment:%d", commentID)
}

// CommentLastActivityPosition returns the 64-bit activity position for the latest state of commentID.
// If the comment was loaded as part of thread comments in the current context, it returns the
// snapshot-consistent activity position that was rendered to the user.
func (cd *CoreData) CommentLastActivityPosition(commentID int32) int64 {
	if cd == nil || commentID <= 0 {
		return CalculateActivityPosition(commentID, 0)
	}
	if cd.cache.forumThreadComments != nil {
		cd.cache.mapMu.Lock()
		for _, lv := range cd.cache.forumThreadComments {
			if lv != nil {
				if comments, ok := lv.Peek(); ok {
					for _, c := range comments {
						if c.Idcomments == commentID {
							cd.cache.mapMu.Unlock()
							return CalculateActivityPosition(c.Idcomments, c.LastAppendID)
						}
					}
				}
			}
		}
		cd.cache.mapMu.Unlock()
	}
	if cd.queries == nil {
		return CalculateActivityPosition(commentID, 0)
	}
	appends, err := cd.queries.SystemListCommentAppendsByCommentID(cd.ctx, commentID)
	if err == nil && len(appends) > 0 {
		lastAppend := appends[len(appends)-1]
		return CalculateActivityPosition(commentID, int64(lastAppend.ID))
	}
	return CalculateActivityPosition(commentID, 0)
}

// CommentDisplayActivity returns the canonical activity ID for a displayed comment row.
func (cd *CoreData) CommentDisplayActivity(cmt *db.GetCommentsByThreadIdForUserRow) string {
	if cmt == nil || cmt.Idcomments <= 0 {
		return ""
	}
	return FormatActivityPosition(cmt.Idcomments, cmt.LastAppendID)
}

// SetThreadReadMarker stores the last read comment for the current user on a thread.
func (cd *CoreData) SetThreadReadMarker(threadID, commentID int32) error {
	return cd.SetThreadReadMarkerPosition(threadID, commentID, CalculateActivityPosition(commentID, 0))
}

// SetThreadReadMarkerPosition stores the last read comment and full activity position watermark
// for the current user on a thread.
func (cd *CoreData) SetThreadReadMarkerPosition(threadID, commentID int32, activityPosition int64) error {
	if cd.queries == nil {
		return nil
	}
	if commentID <= 0 && activityPosition > 0 {
		commentID, _ = ActivityPositionCommentAndAppend(activityPosition)
	}
	if activityPosition <= 0 && commentID > 0 {
		activityPosition = CalculateActivityPosition(commentID, 0)
	}
	return cd.queries.UpsertContentReadMarker(cd.ctx, db.UpsertContentReadMarkerParams{
		Item:                 "thread",
		ItemID:               threadID,
		UserID:               cd.UserID,
		LastCommentID:        commentID,
		LastActivityPosition: activityPosition,
	})
}

// AcknowledgeThreadActivityPosition persists the user's thread read marker and full activity position
// watermark, then marks matching notifications read up to activityPosition.
func (cd *CoreData) AcknowledgeThreadActivityPosition(threadID, commentID int32, activityPosition int64) error {
	return cd.acknowledgeThreadActivityPosition(threadID, commentID, activityPosition, false)
}

func (cd *CoreData) acknowledgeThreadActivityPosition(threadID, commentID int32, activityPosition int64, updateLabels bool) error {
	if cd.queries == nil {
		return nil
	}
	if commentID <= 0 && activityPosition > 0 {
		commentID, _ = ActivityPositionCommentAndAppend(activityPosition)
	}
	if commentID <= 0 {
		return fmt.Errorf("acknowledge thread activity: invalid comment ID %d", commentID)
	}
	if activityPosition <= 0 {
		activityPosition = CalculateActivityPosition(commentID, 0)
	}
	if coordinator, ok := cd.queries.(db.NotificationReadStateCoordinator); ok {
		return coordinator.AcknowledgeThreadReadActivity(cd.ctx, db.AcknowledgeThreadReadActivityParams{
			RecipientID:      cd.UserID,
			ThreadID:         threadID,
			CommentID:        commentID,
			ActivityPosition: activityPosition,
			UpdateLabels:     updateLabels,
		})
	}
	if updateLabels {
		if err := cd.SetThreadPrivateLabelStatus(threadID, false, false); err != nil {
			return fmt.Errorf("set thread private label status: %w", err)
		}
	}
	if err := cd.SetThreadReadMarkerPosition(threadID, commentID, activityPosition); err != nil {
		return fmt.Errorf("save thread read marker: %w", err)
	}
	if err := cd.queries.SystemAcknowledgeNotificationsForResource(cd.ctx, db.SystemAcknowledgeNotificationsForResourceParams{
		RecipientID:      cd.UserID,
		ResourceType:     sql.NullString{String: "thread", Valid: true},
		ResourceID:       sql.NullInt32{Int32: threadID, Valid: true},
		ActivityPosition: sql.NullInt64{Int64: activityPosition, Valid: true},
	}); err != nil {
		return fmt.Errorf("acknowledge thread notifications: %w", err)
	}
	return nil
}

// AcknowledgeThreadActivity persists the user's thread read position and marks matching notifications read.
func (cd *CoreData) AcknowledgeThreadActivity(threadID, commentID int32) error {
	return cd.AcknowledgeThreadActivityPosition(threadID, commentID, CalculateActivityPosition(commentID, 0))
}

// ThreadReadMarker returns the last comment ID read by the current user for a thread.
func (cd *CoreData) ThreadReadMarker(threadID int32) (int32, error) {
	if cd.queries == nil {
		return 0, nil
	}
	row, err := cd.queries.GetContentReadMarker(cd.ctx, db.GetContentReadMarkerParams{
		Item:   "thread",
		ItemID: threadID,
		UserID: cd.UserID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return row.LastCommentID, nil
}

// ThreadReadMarkerPosition returns the last activity position read by the current user for a thread.
func (cd *CoreData) ThreadReadMarkerPosition(threadID int32) (int64, error) {
	if cd.queries == nil {
		return 0, nil
	}
	row, err := cd.queries.GetContentReadMarker(cd.ctx, db.GetContentReadMarkerParams{
		Item:   "thread",
		ItemID: threadID,
		UserID: cd.UserID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	if row.LastActivityPosition > 0 {
		return row.LastActivityPosition, nil
	}
	if row.LastCommentID > 0 {
		return CalculateActivityPosition(row.LastCommentID, 0), nil
	}
	return 0, nil
}
