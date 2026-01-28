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

	// Cutoff: We want to send if last_digest_sent_at is NULL OR older than, say, 20 hours.
	// If it was sent 23 hours ago, it's "yesterday's" digest.
	// If it was sent 1 hour ago, it's "today's".
	cutoff := now.Add(-20 * time.Hour)

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
		// Even if no notifications, we should probably update LastDigestSentAt so we don't keep checking?
		// Or should we only update if we sent an email?
		// Logic: If user has 0 notifications, do we send "You have 0 notifications"?
		// Usually digests are "Here is what you missed". If nothing missed, silence is golden.
		// BUT if we don't update LastDigestSentAt, we will keep checking every 30 mins and finding 0 notifications.
		// This is cheap DB read (index on unread).
		// However, if we don't update timestamp, and then 10 mins later a notification arrives, we WILL send it.
		// That effectively becomes "immediate" delivery for that hour window.
		// That sounds like a feature.
		return nil
	}

	et := NewEmailTemplates("digest")

	baseURL := n.Config.HTTPHostname
	if baseURL == "" {
		baseURL = "https://legacy.arran.net.au"
	}
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
