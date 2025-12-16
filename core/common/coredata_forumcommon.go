package common

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/internal/db"
)

func (cd *CoreData) GrantPrivateForumThread(ctx context.Context, newThreadID int32, userIDs []int32) error {
	for _, userID := range userIDs {
		for _, p := range []string{"see", "view", "post", "reply", "edit"} {
			if _, err := cd.queries.SystemCreateGrant(ctx, db.SystemCreateGrantParams{
				Section: "privateforum",
				Item:    sql.NullString{String: "thread", Valid: true},
				ItemID:  sql.NullInt32{Int32: newThreadID, Valid: true},
				Action:  p,
				UserID:  sql.NullInt32{Int32: userID, Valid: true},
			}); err != nil {
				return fmt.Errorf("granting see on thread %d to user %d: %w", newThreadID, userID, err)
			}
		}
	}
	return nil
}
