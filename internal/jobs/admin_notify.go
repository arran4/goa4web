package jobs

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/opengraph"
)

// jobHistoryPruneInterval is the interval between terminal history cleanup passes.
const jobHistoryPruneInterval = time.Minute

// NewConfiguredService constructs a durable job service wired to notify active administrators
// when a job exhausts its retries.
func NewConfiguredService(q db.Querier) *Service {
	svc := NewService(q)
	svc.recordAdminFailureAlerts = true
	return svc
}

func (s *Service) recordAdminFailureAlertsForJob(ctx context.Context, q db.Querier, job *db.DurableJob, jobErr error, now time.Time) error {
	adminIDs, err := q.SystemListActiveAdminUserIDs(ctx)
	if err != nil {
		return err
	}
	class := adminFailureClass(jobErr)
	identity := sha256.Sum256([]byte(job.DedupeKey + "\x00" + class))
	key := fmt.Sprintf("durable-job-failure:%x", identity)
	summary := fmt.Sprintf("Durable %s job #%d failed (%s)", job.JobType, job.ID, class)
	activityKey := sql.NullString{String: key, Valid: true}
	for _, adminID := range adminIDs {
		if err := q.SystemRecordAdminJobFailureNotification(ctx, db.SystemRecordAdminJobFailureNotificationParams{
			RecipientID: adminID, ActivityKey: activityKey, LatestAt: now,
		}); err != nil {
			return err
		}
		if err := q.SystemSummarizeAdminJobFailureNotification(ctx, db.SystemSummarizeAdminJobFailureNotificationParams{
			RecipientID: adminID, ActivityKey: activityKey, Summary: summary,
			ReopenBefore: sql.NullTime{Time: now.Add(-s.notifyDebounce()), Valid: true},
		}); err != nil {
			return err
		}
	}
	return nil
}

func adminFailureClass(err error) string {
	var status *opengraph.HTTPStatusError
	var networkError net.Error
	var dnsError *net.DNSError
	switch {
	case errors.Is(err, opengraph.ErrInvalidDestination):
		return "invalid destination"
	case errors.As(err, &status):
		return fmt.Sprintf("HTTP %d", status.StatusCode)
	case errors.Is(err, context.DeadlineExceeded), errors.As(err, &networkError) && networkError.Timeout():
		return "timeout"
	case errors.As(err, &dnsError):
		return "DNS failure"
	case networkError != nil:
		return "network failure"
	default:
		return "execution failure"
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
