package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/arran4/goa4web/internal/db"
	domainexternallink "github.com/arran4/goa4web/internal/externallink"
)

// maxDurableJobDedupeKeyRunes matches durable_jobs.dedupe_key VARCHAR(255).
const maxDurableJobDedupeKeyRunes = 255

// AdminEnqueueExternalURL ensures an external_links record exists for canonicalURL
// and atomically enqueues or coalesces a durable external_link job for it.
func (s *Service) AdminEnqueueExternalURL(ctx context.Context, canonicalURL string, actorID sql.NullInt32, force bool) (int32, bool, error) {
	if s == nil || s.Queries == nil {
		return 0, false, nil
	}
	dedupeKey := DedupeKey(JobTypeExternalLink, canonicalURL)
	if utf8.RuneCountInString(dedupeKey) > maxDurableJobDedupeKeyRunes {
		return 0, false, fmt.Errorf("external URL exceeds maximum durable job key length")
	}
	tx, ok := s.Queries.(db.JobTransactor)
	if !ok {
		return 0, false, fmt.Errorf("durable job transactions unavailable")
	}

	var linkID int32
	var created bool
	now := s.now()

	err := tx.WithJobTransaction(ctx, func(q db.Querier) error {
		if force {
			if err := q.SystemClearJobCooldownWindow(ctx, db.SystemClearJobCooldownWindowParams{
				CooldownUntil: now,
				DedupeKey:     dedupeKey,
			}); err != nil {
				return fmt.Errorf("clear cooldown: %w", err)
			}
		} else {
			cd, err := q.SystemGetJobCooldown(ctx, dedupeKey)
			if err == nil && cd.CooldownUntil.After(now) {
				return ErrInCooldown
			}
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("check cooldown: %w", err)
			}
		}

		// Look up before ResolveRecord so existing SQLite rows keep their true ID without ON CONFLICT side effects.
		existing, err := q.GetExternalLink(ctx, canonicalURL)
		switch {
		case err == nil && existing != nil && existing.ID > 0:
			linkID = existing.ID
		case err == nil || errors.Is(err, sql.ErrNoRows):
			resolvedID, resErr := domainexternallink.ResolveRecord(ctx, q, canonicalURL)
			if resErr != nil {
				return fmt.Errorf("resolve external link: %w", resErr)
			}
			linkID = resolvedID
		default:
			return fmt.Errorf("lookup external link: %w", err)
		}
		if linkID == 0 {
			return fmt.Errorf("resolve external link: empty id")
		}

		var enqErr error
		created, enqErr = enqueueActiveJob(ctx, q, EnqueueRequest{
			JobType:         JobTypeExternalLink,
			TargetType:      JobTypeExternalLink,
			TargetID:        int64(linkID),
			TargetKey:       canonicalURL,
			CreatedByUserID: actorID,
		}, dedupeKey, now)
		return enqErr
	})
	if err != nil {
		return 0, false, err
	}
	s.SignalWakeup()
	return linkID, created, nil
}
