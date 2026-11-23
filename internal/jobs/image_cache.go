package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// TargetTypeImageCacheEntry identifies an image job keyed by the persisted cache entry.
const TargetTypeImageCacheEntry = "image_cache_entry"

// imageCacheRecoveryInterval bounds how often workers reconcile pending cache entries.
const imageCacheRecoveryInterval = time.Minute

// QueuePendingImageCacheJobs queues one requested cache entry or a bounded batch
// of overdue entries when entryID is empty. Existing jobs keep their retry state.
func (s *Service) QueuePendingImageCacheJobs(ctx context.Context, entryID string, actorID sql.NullInt32) error {
	now := s.now()
	rows, err := s.Queries.SystemQueuePendingImageCacheJobs(ctx, db.SystemQueuePendingImageCacheJobsParams{
		EntryID:     sql.NullString{String: entryID, Valid: entryID != ""},
		ActorID:     actorID,
		RunAfter:    now,
		DueAt:       sql.NullTime{Time: now, Valid: true},
		CooldownAt:  now,
		RetryFailed: entryID != "",
		MaxAttempts: defaultFetchMaxAttempts,
	})
	if err != nil {
		return fmt.Errorf("queue pending image cache entries: %w", err)
	}
	if rows > 0 {
		s.SignalWakeup()
	}
	return nil
}

// RunImageCacheRecoveryLoop restores image work after an interrupted enqueue.
func (s *Service) RunImageCacheRecoveryLoop(ctx context.Context) {
	ticker := time.NewTicker(imageCacheRecoveryInterval)
	defer ticker.Stop()
	for ctx.Err() == nil {
		if err := s.QueuePendingImageCacheJobs(ctx, "", sql.NullInt32{}); err != nil && ctx.Err() == nil {
			log.Printf("recover image cache jobs: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
