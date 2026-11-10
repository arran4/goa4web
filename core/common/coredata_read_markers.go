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
func (cd *CoreData) CommentLastActivity(commentID int32) string {
	if cd == nil || cd.queries == nil || commentID <= 0 {
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
func (cd *CoreData) CommentLastActivityPosition(commentID int32) int64 {
	if cd == nil || cd.queries == nil || commentID <= 0 {
		return CalculateActivityPosition(commentID, 0)
	}
	appends, err := cd.queries.SystemListCommentAppendsByCommentID(cd.ctx, commentID)
	if err == nil && len(appends) > 0 {
		lastAppend := appends[len(appends)-1]
		return CalculateActivityPosition(commentID, int64(lastAppend.ID))
	}
	return CalculateActivityPosition(commentID, 0)
}

// SetThreadReadMarker stores the last read comment for the current user on a thread.
func (cd *CoreData) SetThreadReadMarker(threadID, commentID int32) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.UpsertContentReadMarker(cd.ctx, db.UpsertContentReadMarkerParams{
		Item:          "thread",
		ItemID:        threadID,
		UserID:        cd.UserID,
		LastCommentID: commentID,
	})
}

// AcknowledgeThreadActivityPosition persists the user's thread read marker and marks matching
// notifications read up to activityPosition.
func (cd *CoreData) AcknowledgeThreadActivityPosition(threadID, commentID int32, activityPosition int64) error {
	if cd.queries == nil {
		return nil
	}
	if commentID <= 0 {
		return fmt.Errorf("acknowledge thread activity: invalid comment ID %d", commentID)
	}
	if err := cd.SetThreadReadMarker(threadID, commentID); err != nil {
		return fmt.Errorf("save thread read marker: %w", err)
	}
	if activityPosition <= 0 {
		activityPosition = CalculateActivityPosition(commentID, 0)
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
