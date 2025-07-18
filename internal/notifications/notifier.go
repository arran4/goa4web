package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/utils/emailutil"
)

// Notifier dispatches updates via email and the internal notification system.
type Notifier struct {
	EmailProvider email.Provider
	Queries       *db.Queries
}

// NotifyChange delivers a single update to the given user and address.
func (n Notifier) NotifyChange(ctx context.Context, userID int32, emailAddr, page, action string, item interface{}) error {
	if n.Queries != nil {
		if err := emailutil.CreateEmailTemplateAndQueue(ctx, n.Queries, userID, emailAddr, page, action, item); err != nil {
			return err
		}
	} else {
		if !emailutil.EmailSendingEnabled() {
			return nil
		}
		msg, toAddr, err := emailutil.CreateEmailTemplate(ctx, emailAddr, page, action, item)
		if err != nil {
			return err
		}
		if n.EmailProvider == nil {
			return fmt.Errorf("no provider")
		}
		if err := n.EmailProvider.Send(ctx, toAddr, msg); err != nil {
			return err
		}
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
	if !emailutil.AdminNotificationsEnabled() {
		return
	}
	if n.EmailProvider == nil && n.Queries == nil {
		return
	}
	for _, addr := range emailutil.GetAdminEmails(ctx, n.Queries) {
		var uid int32
		if n.Queries != nil {
			u, err := n.Queries.UserByEmail(ctx, addr)
			if err != nil {
				log.Printf("user by email %s: %v", addr, err)
			} else {
				uid = u.Idusers
			}
		}
		if err := n.NotifyChange(ctx, uid, addr, page, "update", nil); err != nil {
			log.Printf("notify admin %s: %v", addr, err)
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
		ForumthreadID: threadID,
		Idusers:       excludeUser,
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

// NotifyWritingSubscribers informs subscribed users about a writing update.
func (n Notifier) NotifyWritingSubscribers(ctx context.Context, writingID, excludeUser int32, page string) {
	emailutil.NotifyWritingSubscribers(ctx, n.Queries, writingID, excludeUser, page)
}
