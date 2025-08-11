package common

import (
	"database/sql"
	"errors"

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
