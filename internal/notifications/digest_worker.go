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
	now := time.Now().UTC()

	// Cutoff: We want to send if last_digest_sent_at is NULL OR older than 24 hours.
	cutoff := now.Add(-24 * time.Hour)

	// 1. Process users with no timezone (assume UTC)
	utcHour := now.Hour()
	usersNoTz, err := n.Queries.GetUsersForDailyDigestNoTimezone(ctx, db.GetUsersForDailyDigestNoTimezoneParams{
		Hour:   sql.NullInt32{Int32: int32(utcHour), Valid: true},
		Cutoff: sql.NullTime{Time: cutoff, Valid: true},
	})
	if err != nil {
		log.Printf("GetUsersForDailyDigestNoTimezone: %v", err)
	} else {
		for _, user := range usersNoTz {
			if err := n.sendDigestToUser(ctx, user.UsersIdusers, user.Email, user.DailyDigestMarkRead); err != nil {
				log.Printf("sendDigestToUser (no tz) %d: %v", user.UsersIdusers, err)
			}
		}
	}

	// 2. Process users with specific timezones
	timezones, err := n.Queries.GetDigestTimezones(ctx)
	if err != nil {
		log.Printf("GetDigestTimezones: %v", err)
		return
	}

	for _, tzNullStr := range timezones {
		if !tzNullStr.Valid || tzNullStr.String == "" {
			continue // Should be handled by NoTimezone query, but safe to skip
		}
		tzStr := tzNullStr.String
		loc, err := time.LoadLocation(tzStr)
		if err != nil {
			log.Printf("Invalid timezone %s: %v", tzStr, err)
			continue
		}

		localHour := now.In(loc).Hour()
		users, err := n.Queries.GetUsersForDailyDigestByTimezone(ctx, db.GetUsersForDailyDigestByTimezoneParams{
			Hour:     sql.NullInt32{Int32: int32(localHour), Valid: true},
			Timezone: tzNullStr,
			Cutoff:   sql.NullTime{Time: cutoff, Valid: true},
		})
		if err != nil {
			log.Printf("GetUsersForDailyDigestByTimezone (%s): %v", tzStr, err)
			continue
		}

		for _, user := range users {
			if err := n.sendDigestToUser(ctx, user.UsersIdusers, user.Email, user.DailyDigestMarkRead); err != nil {
				log.Printf("sendDigestToUser (%s) %d: %v", tzStr, user.UsersIdusers, err)
			}
		}
	}
}

func (n *Notifier) sendDigestToUser(ctx context.Context, userID int32, email string, markRead bool) error {
	// Get unread notifications
	limit := int32(2147483647) // Max Int32
	notifs, err := n.Queries.ListUnreadNotificationsForLister(ctx, db.ListUnreadNotificationsForListerParams{
		ListerID: userID,
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

	if err := n.renderAndQueueEmailFromTemplates(ctx, &userID, email, et, data); err != nil {
		return err
	}

	// Update last sent time
	if err := n.Queries.UpdateLastDigestSentAt(ctx, db.UpdateLastDigestSentAtParams{
		SentAt:   sql.NullTime{Time: time.Now().UTC(), Valid: true},
		ListerID: userID,
	}); err != nil {
		log.Printf("UpdateLastDigestSentAt: %v", err)
	}

	if markRead {
		ids := make([]int32, len(notifs))
		for i, n := range notifs {
			ids[i] = n.ID
		}
		if err := n.Queries.SetNotificationsReadForListerBatch(ctx, db.SetNotificationsReadForListerBatchParams{
			ListerID: userID,
			Ids:      ids,
		}); err != nil {
			log.Printf("Mark read error: %v", err)
		}
	}
	return nil
}
