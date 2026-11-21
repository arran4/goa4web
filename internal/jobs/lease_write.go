package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// ErrLeaseLost means the worker's claim expired, was replaced, or was cancelled.
var ErrLeaseLost = errors.New("durable job lease lost")

// WriteWithActiveLease checks the current claim and commits its database updates
// together. Network and storage operations belong outside this transaction.
func (s *Service) WriteWithActiveLease(ctx context.Context, job *db.DurableJob, write func(db.Querier) error) error {
	tx, ok := s.Queries.(db.JobTransactor)
	if !ok {
		return fmt.Errorf("durable job transactions unavailable")
	}
	return tx.WithJobTransaction(ctx, func(q db.Querier) error {
		_, err := q.SystemLockActiveDurableJobLease(ctx, db.SystemLockActiveDurableJobLeaseParams{
			ID: job.ID, LeaseToken: job.LeaseToken,
			NowAt: sql.NullTime{Time: s.now(), Valid: true},
		})
		if errors.Is(err, sql.ErrNoRows) {
			return ErrLeaseLost
		}
		if err != nil {
			return fmt.Errorf("check durable job lease: %w", err)
		}
		return write(q)
	})
}
