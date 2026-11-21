package common

import (
	"context"
	"database/sql"
	"slices"
	"time"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
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
	if err := cd.queries.SetNotificationReadForLister(ctx, db.SetNotificationReadForListerParams{
		ID:       notificationID,
		ListerID: cd.UserID,
	}); err != nil {
		return err
	}
	cd.publishNotificationChange()
	return nil
}

// DismissNotificationSnapshot acknowledges exactly the listed notifications
// owned by the current actor. It leaves the thread watermark untouched.
func (cd *CoreData) DismissNotificationSnapshot(ctx context.Context, ids []int32) error {
	if cd == nil || cd.queries == nil || cd.UserID <= 0 || len(ids) == 0 {
		return nil
	}
	// Keep each bind list within both database engines' parameter limits.
	for batch := range slices.Chunk(ids, 500) {
		if err := cd.queries.SetNotificationsReadForListerBatch(ctx, db.SetNotificationsReadForListerBatchParams{
			ListerID: cd.UserID, Ids: batch,
		}); err != nil {
			cd.publishNotificationChange()
			return err
		}
	}
	cd.publishNotificationChange()
	return nil
}

func (cd *CoreData) publishNotificationChange() {
	if cd.EventBus() != nil {
		_ = cd.Publish(eventbus.NotificationsChanged{RecipientID: cd.UserID, Time: time.Now()})
	}
}
