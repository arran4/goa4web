package jobs

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// jobHistoryPruneInterval is the interval between terminal history cleanup passes.
const jobHistoryPruneInterval = time.Minute

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

// RunPruneLoop removes excess terminal job history at startup and periodically.
func (s *Service) RunPruneLoop(ctx context.Context, maxAge time.Duration, maxPerKind int) {
	ticker := time.NewTicker(jobHistoryPruneInterval)
	defer ticker.Stop()
	for ctx.Err() == nil {
		if _, err := s.PruneTerminal(ctx, maxAge, maxPerKind); err != nil && ctx.Err() == nil {
			log.Printf("prune terminal durable jobs: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
