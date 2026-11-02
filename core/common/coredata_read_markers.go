package common

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

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

// AcknowledgeThreadActivity persists the user's thread read position and marks matching notifications read.
func (cd *CoreData) AcknowledgeThreadActivity(threadID, commentID int32) error {
	if cd.queries == nil {
		return nil
	}
	if commentID <= 0 {
		return fmt.Errorf("acknowledge thread activity: invalid comment ID %d", commentID)
	}
	if err := cd.SetThreadReadMarker(threadID, commentID); err != nil {
		return fmt.Errorf("save thread read marker: %w", err)
	}
	if err := cd.queries.SystemAcknowledgeNotificationsForResource(cd.ctx, db.SystemAcknowledgeNotificationsForResourceParams{
		RecipientID:      cd.UserID,
		ResourceType:     sql.NullString{String: "thread", Valid: true},
		ResourceID:       sql.NullInt32{Int32: threadID, Valid: true},
		ActivityPosition: sql.NullInt32{Int32: commentID, Valid: true},
	}); err != nil {
		return fmt.Errorf("acknowledge thread notifications: %w", err)
	}
	return nil
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
