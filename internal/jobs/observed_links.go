package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"slices"
	"unicode/utf8"

	"github.com/arran4/goa4web/internal/db"
)

// QueueObservedExternalLink queues a link found in content. An already active
// job retains its retry schedule and an existing cooldown still applies.
func (s *Service) QueueObservedExternalLink(ctx context.Context, target string, actorID sql.NullInt32) error {
	queued, err := s.RecordObservedExternalLinks(ctx, s.Queries, []string{target}, actorID)
	if queued {
		s.SignalWakeup()
	}
	return err
}

// RecordObservedExternalLinks records jobs using the caller's querier. Callers
// supplying a transaction signal the worker only after commit.
func (s *Service) RecordObservedExternalLinks(ctx context.Context, q db.Querier, targets []string, actorID sql.NullInt32) (bool, error) {
	targets = slices.Clone(targets)
	slices.Sort(targets)
	now := s.now()
	queued := false
	for _, target := range slices.Compact(targets) {
		destination, err := url.Parse(target)
		if err != nil || destination.Hostname() == "" || (destination.Scheme != "http" && destination.Scheme != "https") {
			continue
		}
		key := DedupeKey(JobTypeExternalLink, target)
		// durableJobDedupeKeyLength is the number of characters stored by the queue's dedupe column.
		const durableJobDedupeKeyLength = 255
		if utf8.RuneCountInString(key) > durableJobDedupeKeyLength {
			continue
		}
		rows, err := q.SystemQueueObservedExternalLinkJob(ctx, db.SystemQueueObservedExternalLinkJobParams{
			DedupeKey:   key,
			ActiveKey:   sql.NullString{String: key, Valid: true},
			TargetKey:   target,
			MaxAttempts: defaultFetchMaxAttempts,
			RunAfter:    now,
			ActorID:     actorID,
			CheckKey:    key,
			CooldownAt:  now,
		})
		if err != nil {
			return queued, fmt.Errorf("queue observed external link: %w", err)
		}
		queued = queued || rows > 0
	}
	return queued, nil
}
