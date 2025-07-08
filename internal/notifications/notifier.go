package notifications

import (
	"context"
	"database/sql"
	"log"

	"github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/emailutil"
)

// Notifier dispatches updates via email and the internal notification system.
type Notifier struct {
	EmailProvider email.Provider
	Queries       *db.Queries
}

// NotifyChange delivers a single update to the given user and address.
func (n Notifier) NotifyChange(ctx context.Context, userID int32, emailAddr, page, action string, item interface{}) error {
	if err := emailutil.NotifyChange(ctx, n.EmailProvider, userID, emailAddr, page, action, item); err != nil {
		return err
	}
	if n.Queries != nil && common.NotificationsEnabled() && userID != 0 {
		err := n.Queries.InsertNotification(ctx, db.InsertNotificationParams{
			UsersIdusers: userID,
			Link:         sql.NullString{String: page, Valid: page != ""},
			Message:      sql.NullString{},
		})
		if err != nil {
			log.Printf("insert notification: %v", err)
		}
	}
	return nil
}

// NotifyAdmins sends a change notification to all administrator accounts.
func (n Notifier) NotifyAdmins(ctx context.Context, page string) {
	emailutil.NotifyAdmins(ctx, n.EmailProvider, n.Queries, page)
	if n.Queries == nil || !common.NotificationsEnabled() {
		return
	}
	for _, addr := range emailutil.GetAdminEmails(ctx, n.Queries) {
		u, err := n.Queries.UserByEmail(ctx, sql.NullString{String: addr, Valid: true})
		if err != nil {
			log.Printf("user by email %s: %v", addr, err)
			continue
		}
		if err := n.Queries.InsertNotification(ctx, db.InsertNotificationParams{
			UsersIdusers: u.Idusers,
			Link:         sql.NullString{String: page, Valid: page != ""},
			Message:      sql.NullString{},
		}); err != nil {
			log.Printf("insert notification: %v", err)
		}
	}
}

// NotifyThreadSubscribers informs subscribed users about a thread update.
func (n Notifier) NotifyThreadSubscribers(ctx context.Context, threadID, excludeUser int32, page string) {
	emailutil.NotifyThreadSubscribers(ctx, n.EmailProvider, n.Queries, threadID, excludeUser, page)
	if n.Queries == nil || !common.NotificationsEnabled() {
		return
	}
	rows, err := n.Queries.ListUsersSubscribedToThread(ctx, db.ListUsersSubscribedToThreadParams{
		ForumthreadIdforumthread: threadID,
		Idusers:                  excludeUser,
	})
	if err != nil {
		log.Printf("list subscribers: %v", err)
		return
	}
	for _, row := range rows {
		if err := n.Queries.InsertNotification(ctx, db.InsertNotificationParams{
			UsersIdusers: row.UsersIdusers,
			Link:         sql.NullString{String: page, Valid: page != ""},
			Message:      sql.NullString{},
		}); err != nil {
			log.Printf("insert notification: %v", err)
		}
	}
}
