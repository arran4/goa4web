package main

import (
	"context"
	"database/sql"
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

// InsertNotification stores a notification for the user.
type InsertNotificationParams struct {
	UsersIdusers int32
	Link         sql.NullString
	Message      sql.NullString
}

func (q *Queries) InsertNotification(ctx context.Context, arg InsertNotificationParams) error {
	_, err := q.db.ExecContext(ctx,
		"INSERT INTO notifications (users_idusers, link, message) VALUES (?, ?, ?)",
		arg.UsersIdusers, arg.Link, arg.Message)
	return err
}

// CountUnreadNotifications returns the number of unread notifications for a user.
func (q *Queries) CountUnreadNotifications(ctx context.Context, userID int32) (int32, error) {
	var c int32
	err := q.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM notifications WHERE users_idusers = ? AND read_at IS NULL",
		userID).Scan(&c)
	return c, err
}

// GetUnreadNotifications returns unread notifications for the user ordered newest first.
func (q *Queries) GetUnreadNotifications(ctx context.Context, userID int32) ([]*Notification, error) {
	rows, err := q.db.QueryContext(ctx,
		"SELECT id, users_idusers, link, message, created_at, read_at FROM notifications WHERE users_idusers = ? AND read_at IS NULL ORDER BY id DESC",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UsersIdusers, &n.Link, &n.Message, &n.CreatedAt, &n.ReadAt); err != nil {
			return nil, err
		}
		items = append(items, &n)
	}
	return items, rows.Err()
}

// MarkNotificationRead marks the notification as read now.
func (q *Queries) MarkNotificationRead(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, "UPDATE notifications SET read_at = NOW() WHERE id = ?", id)
	return err
}

// PurgeReadNotifications deletes notifications read over 24h ago.
func (q *Queries) PurgeReadNotifications(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx,
		"DELETE FROM notifications WHERE read_at IS NOT NULL AND read_at < (NOW() - INTERVAL 24 HOUR)")
	return err
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
