package notifications

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	db "github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/feeds"
)

// NotificationsFeed produces a feed from the notifications slice.
func NotificationsFeed(r *http.Request, notifications []*db.Notification) *feeds.Feed {
	feed := &feeds.Feed{
		Title:       "Notifications",
		Link:        &feeds.Link{Href: r.URL.Path},
		Description: "recent notifications",
		Created:     time.Now(),
	}
	for _, n := range notifications {
		link := ""
		if n.Link.Valid {
			link = n.Link.String
		}
		msg := ""
		if n.Message.Valid {
			msg = n.Message.String
		}
		item := &feeds.Item{
			Title:       msg,
			Link:        &feeds.Link{Href: link},
			Created:     time.Now(),
			Description: msg,
		}
		item.Created = n.CreatedAt
		feed.Items = append(feed.Items, item)
	}
	return feed
}

// notificationPurgeWorker periodically removes old read notifications.
func (n *Notifier) NotificationPurgeWorker(ctx context.Context, interval time.Duration) {
	if n.Queries == nil {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := n.Queries.PurgeReadNotifications(ctx); err != nil {
				log.Printf("purge notifications: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// sendInternalNotification stores an internal notification for the user.
func sendInternalNotification(ctx context.Context, q *db.Queries, userID int32, path, msg string) error {
	return q.InsertNotification(ctx, db.InsertNotificationParams{
		UsersIdusers: userID,
		Link:         sql.NullString{String: path, Valid: path != ""},
		Message:      sql.NullString{String: msg, Valid: msg != ""},
	})
}
