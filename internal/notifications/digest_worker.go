package notifications

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

type DigestType int

const (
	DigestDaily DigestType = iota
	DigestWeekly
	DigestMonthly
)

const SchedulerTaskName = "digest_scheduler"

// ScheduleDigest publishes a digest run event for the given time.
func (n *Notifier) ScheduleDigest(ctx context.Context, t time.Time) error {
	if n.Bus == nil {
		return nil
	}
	evt := eventbus.DigestRunEvent{Time: t}
	return n.Bus.Publish(evt)
}

// ProcessDigestForTime is called by the consumer to process digests for a specific time.
// It was previously private `processDigestForTime`.
func (n *Notifier) ProcessDigestForTime(ctx context.Context, t time.Time) {
	log.Printf("Processing digests for time: %v", t)

	// Cutoff times: send if not sent since cutoff.
	// For Daily: 23 hours ago (allow some buffer)
	dailyCutoff := t.Add(-23 * time.Hour)
	// For Weekly: 6 days 23 hours ago
	weeklyCutoff := t.Add(-24 * 7 * time.Hour).Add(time.Hour)
	// For Monthly: 27 days ago (rough approx, or just ensure distinct month)
	monthlyCutoff := t.Add(-24 * 27 * time.Hour)

	// 1. Daily Digests
	// UTC
	users, err := n.Queries.GetUsersForDailyDigestNoTimezone(ctx, db.GetUsersForDailyDigestNoTimezoneParams{
		Hour:   sql.NullInt32{Int32: int32(t.Hour()), Valid: true},
		Cutoff: sql.NullTime{Time: dailyCutoff, Valid: true},
	})
	if err != nil {
		log.Printf("GetUsersForDailyDigestNoTimezone: %v", err)
	}
	for _, u := range users {
		n.safeSendDigest(ctx, u.UsersIdusers, u.Email, u.DailyDigestMarkRead, DigestDaily)
	}

	// Timezone
	tzs, err := n.Queries.GetDigestTimezones(ctx)
	if err == nil {
		for _, tzNull := range tzs {
			if !tzNull.Valid || tzNull.String == "" {
				continue
			}
			loc, err := time.LoadLocation(tzNull.String)
			if err != nil {
				continue
			}
			localTime := t.In(loc)
			users, err := n.Queries.GetUsersForDailyDigestByTimezone(ctx, db.GetUsersForDailyDigestByTimezoneParams{
				Hour:     sql.NullInt32{Int32: int32(localTime.Hour()), Valid: true},
				Timezone: tzNull,
				Cutoff:   sql.NullTime{Time: dailyCutoff, Valid: true},
			})
			if err != nil {
				log.Printf("GetUsersForDailyDigestByTimezone(%s): %v", tzNull.String, err)
				continue
			}
			for _, u := range users {
				n.safeSendDigest(ctx, u.UsersIdusers, u.Email, u.DailyDigestMarkRead, DigestDaily)
			}
		}
	}

	// 2. Weekly Digests
	// UTC
	// t.Weekday(): Sunday=0, Monday=1...
	usersW, err := n.Queries.GetUsersForWeeklyDigestNoTimezone(ctx, db.GetUsersForWeeklyDigestNoTimezoneParams{
		Day:    sql.NullInt32{Int32: int32(t.Weekday()), Valid: true},
		Hour:   sql.NullInt32{Int32: int32(t.Hour()), Valid: true},
		Cutoff: sql.NullTime{Time: weeklyCutoff, Valid: true},
	})
	if err != nil {
		log.Printf("GetUsersForWeeklyDigestNoTimezone: %v", err)
	}
	for _, u := range usersW {
		n.safeSendDigest(ctx, u.UsersIdusers, u.Email, u.DailyDigestMarkRead, DigestWeekly)
	}

	// Timezone
	if err == nil { // Reuse tzs
		for _, tzNull := range tzs {
			if !tzNull.Valid || tzNull.String == "" {
				continue
			}
			loc, err := time.LoadLocation(tzNull.String)
			if err != nil {
				continue
			}
			localTime := t.In(loc)
			users, err := n.Queries.GetUsersForWeeklyDigestByTimezone(ctx, db.GetUsersForWeeklyDigestByTimezoneParams{
				Day:      sql.NullInt32{Int32: int32(localTime.Weekday()), Valid: true},
				Hour:     sql.NullInt32{Int32: int32(localTime.Hour()), Valid: true},
				Timezone: tzNull,
				Cutoff:   sql.NullTime{Time: weeklyCutoff, Valid: true},
			})
			if err != nil {
				log.Printf("GetUsersForWeeklyDigestByTimezone(%s): %v", tzNull.String, err)
				continue
			}
			for _, u := range users {
				n.safeSendDigest(ctx, u.UsersIdusers, u.Email, u.DailyDigestMarkRead, DigestWeekly)
			}
		}
	}

	// 3. Monthly Digests
	// Handle end-of-month catch-up (e.g. run 31st configs on 28th/30th if applicable)
	// We check the current day 'd'. If 'd' is the last day of the month, we also check d+1..31.

	// Helper to process a specific day config
	processMonthlyDay := func(day int, hour int, tz sql.NullString) {
		var users []*db.GetUsersForMonthlyDigestNoTimezoneRow
		var err error

		if !tz.Valid || tz.String == "" {
			users, err = n.Queries.GetUsersForMonthlyDigestNoTimezone(ctx, db.GetUsersForMonthlyDigestNoTimezoneParams{
				Day:    sql.NullInt32{Int32: int32(day), Valid: true},
				Hour:   sql.NullInt32{Int32: int32(hour), Valid: true},
				Cutoff: sql.NullTime{Time: monthlyCutoff, Valid: true},
			})
			if err != nil {
				log.Printf("GetUsersForMonthlyDigestNoTimezone(d=%d): %v", day, err)
				return
			}
		} else {
			// Map to ByTimezone struct (same structure)
			var usersTz []*db.GetUsersForMonthlyDigestByTimezoneRow
			usersTz, err = n.Queries.GetUsersForMonthlyDigestByTimezone(ctx, db.GetUsersForMonthlyDigestByTimezoneParams{
				Day:      sql.NullInt32{Int32: int32(day), Valid: true},
				Hour:     sql.NullInt32{Int32: int32(hour), Valid: true},
				Timezone: tz,
				Cutoff:   sql.NullTime{Time: monthlyCutoff, Valid: true},
			})
			if err != nil {
				log.Printf("GetUsersForMonthlyDigestByTimezone(%s, d=%d): %v", tz.String, day, err)
				return
			}
			// Convert to shared type or just loop
			for _, u := range usersTz {
				n.safeSendDigest(ctx, u.UsersIdusers, u.Email, u.DailyDigestMarkRead, DigestMonthly)
			}
			return
		}

		for _, u := range users {
			n.safeSendDigest(ctx, u.UsersIdusers, u.Email, u.DailyDigestMarkRead, DigestMonthly)
		}
	}

	// UTC
	currentDay := t.Day()
	processMonthlyDay(currentDay, t.Hour(), sql.NullString{})

	// Check if today is the last day of the UTC month
	year, month, _ := t.Date()
	lastDayOfMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, t.Location()).Day()
	if currentDay == lastDayOfMonth {
		for d := currentDay + 1; d <= 31; d++ {
			processMonthlyDay(d, t.Hour(), sql.NullString{})
		}
	}

	// Timezone
	if err == nil {
		for _, tzNull := range tzs {
			if !tzNull.Valid || tzNull.String == "" {
				continue
			}
			loc, err := time.LoadLocation(tzNull.String)
			if err != nil {
				continue
			}
			localTime := t.In(loc)
			lDay := localTime.Day()
			processMonthlyDay(lDay, localTime.Hour(), tzNull)

			// Check last day of month for this timezone
			lYear, lMonth, _ := localTime.Date()
			lLastDay := time.Date(lYear, lMonth+1, 0, 0, 0, 0, 0, loc).Day()
			if lDay == lLastDay {
				for d := lDay + 1; d <= 31; d++ {
					processMonthlyDay(d, localTime.Hour(), tzNull)
				}
			}
		}
	}
}

func (n *Notifier) safeSendDigest(ctx context.Context, userID int32, email string, markRead bool, dtype DigestType) {
	if err := n.SendDigestToUser(ctx, userID, email, markRead, false, dtype); err != nil {
		log.Printf("Error sending digest type %d to user %d: %v", dtype, userID, err)
	}
}

// SendDigestToUser sends a digest email.
func (n *Notifier) SendDigestToUser(ctx context.Context, userID int32, email string, markRead, preview bool, dtype DigestType) error {
	limit := int32(2147483647)
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

	digestTitle := "Daily Digest"
	switch dtype {
	case DigestWeekly:
		digestTitle = "Weekly Digest"
	case DigestMonthly:
		digestTitle = "Monthly Digest"
	}

	data := struct {
		Notifications []*db.Notification
		BaseURL       string
		DigestTitle   string
	}{
		Notifications: notifs,
		BaseURL:       baseURL,
		DigestTitle:   digestTitle,
	}

	if err := n.renderAndQueueEmailFromTemplates(ctx, &userID, email, et, data); err != nil {
		return err
	}

	if !preview {
		now := time.Now().UTC()
		// Update last sent time based on type
		switch dtype {
		case DigestDaily:
			if err := n.Queries.UpdateLastDigestSentAt(ctx, db.UpdateLastDigestSentAtParams{
				SentAt:   sql.NullTime{Time: now, Valid: true},
				ListerID: userID,
			}); err != nil {
				log.Printf("UpdateLastDigestSentAt: %v", err)
			}
		case DigestWeekly:
			if err := n.Queries.UpdateLastWeeklyDigestSentAt(ctx, db.UpdateLastWeeklyDigestSentAtParams{
				SentAt:   sql.NullTime{Time: now, Valid: true},
				ListerID: userID,
			}); err != nil {
				log.Printf("UpdateLastWeeklyDigestSentAt: %v", err)
			}
		case DigestMonthly:
			if err := n.Queries.UpdateLastMonthlyDigestSentAt(ctx, db.UpdateLastMonthlyDigestSentAtParams{
				SentAt:   sql.NullTime{Time: now, Valid: true},
				ListerID: userID,
			}); err != nil {
				log.Printf("UpdateLastMonthlyDigestSentAt: %v", err)
			}
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
	}
	return nil
}
