package goa4web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	config "github.com/arran4/goa4web/config"
	"github.com/gorilla/feeds"
)

// notificationsEnabled reports if the internal notification system should run.
func notificationsEnabled() bool {
	v := strings.ToLower(os.Getenv(config.EnvNotificationsEnabled))
	if v == "" {
		return true
	}
	switch v {
	case "0", "false", "off", "no":
		return false
	default:
		return true
	}
}

// notificationsFeed produces a feed from the notifications slice.
func notificationsFeed(r *http.Request, notifications []*Notification) *feeds.Feed {
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
func notificationPurgeWorker(ctx context.Context, q *Queries, interval time.Duration) {
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
