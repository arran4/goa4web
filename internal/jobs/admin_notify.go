package jobs

import (
	"context"
	"database/sql"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

const (
	// defaultPruneInterval controls how often terminal durable jobs are pruned.
	defaultPruneInterval = 6 * time.Hour
	// defaultPruneMaxAge controls how long completed/failed/cancelled durable jobs are retained.
	defaultPruneMaxAge = 7 * 24 * time.Hour
)

// NewConfiguredService constructs a durable job service wired to notify active administrators
// when a job exhausts its retries.
func NewConfiguredService(q db.Querier) *Service {
	svc := NewService(q)
	svc.NotifyAdmins = NewAdminNotifier(q)
	return svc
}

// NewAdminNotifier returns a callback that creates an unread site notification for each active admin user.
func NewAdminNotifier(q db.Querier) func(ctx context.Context, message, link string) error {
	return func(ctx context.Context, message, link string) error {
		if q == nil {
			return nil
		}
		adminIDs, err := q.SystemListActiveAdminUserIDs(ctx)
		if err != nil {
			return err
		}
		for _, adminID := range adminIDs {
			if err := q.SystemCreateNotification(ctx, db.SystemCreateNotificationParams{
				RecipientID: adminID,
				Link:        sql.NullString{String: link, Valid: link != ""},
				Message:     sql.NullString{String: message, Valid: message != ""},
			}); err != nil {
				return err
			}
		}
		return nil
	}
}

// RunPruneLoop periodically removes old terminal durable jobs while preserving cooldowns.
func (s *Service) RunPruneLoop(ctx context.Context, interval, maxAge time.Duration) {
	if s == nil {
		return
	}
	if interval <= 0 {
		interval = defaultPruneInterval
	}
	if maxAge <= 0 {
		maxAge = defaultPruneMaxAge
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = s.PruneTerminal(ctx, maxAge)
		}
	}
}
