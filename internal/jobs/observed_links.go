package jobs

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// QueueObservedExternalLink queues a link found in content. An already active
// job retains its retry schedule and an existing cooldown still applies.
func (s *Service) QueueObservedExternalLink(ctx context.Context, url string, actorID sql.NullInt32) error {
	now := s.now()
	key := DedupeKey(JobTypeExternalLink, url)
	rows, err := s.Queries.SystemQueueObservedExternalLinkJob(ctx, db.SystemQueueObservedExternalLinkJobParams{
		DedupeKey:   key,
		ActiveKey:   sql.NullString{String: key, Valid: true},
		TargetKey:   url,
		MaxAttempts: defaultFetchMaxAttempts,
		RunAfter:    now,
		ActorID:     actorID,
		CheckKey:    key,
		CooldownAt:  now,
	})
	if err != nil {
		return fmt.Errorf("queue observed external link: %w", err)
	}
	if rows > 0 {
		s.SignalWakeup()
	}
	return nil
}
