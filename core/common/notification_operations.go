package common

import (
	"context"
	"database/sql"

	"github.com/arran4/goa4web/internal/db"
)

// NotificationIDByActivity returns the actor's notification ID for one activity.
func (cd *CoreData) NotificationIDByActivity(ctx context.Context, activityKey string) (int32, error) {
	if cd == nil || cd.queries == nil || cd.UserID <= 0 || activityKey == "" {
		return 0, sql.ErrNoRows
	}
	return cd.queries.GetNotificationIDByActivityForLister(ctx, db.GetNotificationIDByActivityForListerParams{
		ListerID:    cd.UserID,
		ActivityKey: sql.NullString{String: activityKey, Valid: true},
	})
}

// DismissNotification acknowledges a notification owned by the actor.
// It does not update the content read position of the associated resource.
func (cd *CoreData) DismissNotification(ctx context.Context, notificationID int32) error {
	if cd == nil || cd.queries == nil || cd.UserID <= 0 || notificationID <= 0 {
		return nil
	}
	return cd.queries.SetNotificationReadForLister(ctx, db.SetNotificationReadForListerParams{
		ID:       notificationID,
		ListerID: cd.UserID,
	})
}
