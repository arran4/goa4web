package notifications

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// NotificationDigestWorker runs periodically to send daily digests.
func (n *Notifier) NotificationDigestWorker(ctx context.Context, interval time.Duration) {
	if n.Queries == nil {
		return
	}
	// Initial check on startup (optional, but good for testing if intervals are long)
	// safeGo in workers.go handles panics.
	// n.processDigests(ctx) // Uncomment if we want immediate run

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			n.processDigests(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (n *Notifier) processDigests(ctx context.Context) {
	// Calculate current hour (UTC)
	now := time.Now().UTC()
	currentHour := now.Hour()

	// Cutoff: We want to send if last_digest_sent_at is NULL OR older than 24 hours.
	// We use 24h to strictly enforce a daily cycle.
	cutoff := now.Add(-24 * time.Hour)

	// Get users configured for this hour who haven't received a digest recently
	users, err := n.Queries.GetUsersForDailyDigest(ctx, db.GetUsersForDailyDigestParams{
		Hour:   sql.NullInt32{Int32: int32(currentHour), Valid: true},
		Cutoff: sql.NullTime{Time: cutoff, Valid: true},
	})
	if err != nil {
		log.Printf("GetUsersForDailyDigest: %v", err)
		return
	}

	for _, user := range users {
		if err := n.sendDigestToUser(ctx, user); err != nil {
			log.Printf("sendDigestToUser %d: %v", user.UsersIdusers, err)
		}
	}
}

func (n *Notifier) sendDigestToUser(ctx context.Context, user *db.GetUsersForDailyDigestRow) error {
	// Get unread notifications
	limit := int32(2147483647) // Max Int32
	notifs, err := n.Queries.ListUnreadNotificationsForLister(ctx, db.ListUnreadNotificationsForListerParams{
		ListerID: user.UsersIdusers,
		Limit:    limit,
		Offset:   0,
	})
	if err != nil {
		return err
	}
	if len(notifs) == 0 {
		return nil
	}

	et := NewEmailTemplates("digest")

	baseURL := n.Config.BaseURL
	baseURL = strings.TrimRight(baseURL, "/")

	data := struct {
		Notifications []*db.Notification
		BaseURL       string
	}{
		Notifications: notifs,
		BaseURL:       baseURL,
	}

	if err := n.renderAndQueueEmailFromTemplates(ctx, &user.UsersIdusers, user.Email, et, data); err != nil {
		return err
	}

	// Update last sent time
	if err := n.Queries.UpdateLastDigestSentAt(ctx, db.UpdateLastDigestSentAtParams{
		SentAt:   sql.NullTime{Time: time.Now().UTC(), Valid: true},
		ListerID: user.UsersIdusers,
	}); err != nil {
		log.Printf("UpdateLastDigestSentAt: %v", err)
	}

	if user.DailyDigestMarkRead {
		ids := make([]int32, len(notifs))
		for i, n := range notifs {
			ids[i] = n.ID
		}
		if err := n.Queries.SetNotificationsReadForListerBatch(ctx, db.SetNotificationsReadForListerBatchParams{
			ListerID: user.UsersIdusers,
			Ids:      ids,
		}); err != nil {
			log.Printf("Mark read error: %v", err)
		}
	}
	return nil
}
