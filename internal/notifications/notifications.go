package notifications

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/feeds"

	"github.com/arran4/goa4web/internal/db"
)

// Feed produces a feed from the notifications slice.
func Feed(r *http.Request, notifications []*db.Notification) *feeds.Feed {
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

// PurgeWorker periodically removes old read notifications.
func PurgeWorker(ctx context.Context, q *db.Queries, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := q.PurgeReadNotifications(ctx); err != nil {
				fmt.Println("purge notifications:", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
