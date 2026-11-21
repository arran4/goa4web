// Package jobs implements DB-backed durable background job queueing, leasing, fencing, cooldowns, and worker loops.
package jobs

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

const (
	// JobTypeExternalLink identifies durable external-link metadata and OpenGraph fetch jobs.
	JobTypeExternalLink = "external_link"
	// JobTypeImageCache identifies durable remote image fetch and thumbnail cache jobs.
	JobTypeImageCache = "image_cache"
	// JobTypeSearchIndex identifies durable search-index refresh jobs backed by dirty revisions.
	JobTypeSearchIndex = "search_index"

	// StatusPending indicates a queued job waiting for its run_after timestamp.
	StatusPending = "pending"
	// StatusRunning indicates a job currently leased by a worker.
	StatusRunning = "running"
	// StatusCompleted indicates a job that finished successfully.
	StatusCompleted = "completed"
	// StatusFailed indicates a job that exhausted its retry budget.
	StatusFailed = "failed"
	// StatusCancelled indicates a job cancelled by an administrator.
	StatusCancelled = "cancelled"

	// DefaultCooldown is the default cooldown applied after permanent exhaustion of transient failures.
	DefaultCooldown = 24 * time.Hour
	// DefaultLeaseDuration is the default worker claim lease duration before expiry reclamation.
	DefaultLeaseDuration = 2 * time.Minute
	// DefaultPollInterval is the fallback poll interval if a wakeup signal is missed.
	DefaultPollInterval = 2 * time.Second
	// DefaultMaxAttempts is the default attempt limit before marking a job failed.
	DefaultMaxAttempts = 5
	// defaultFetchMaxAttempts allows an initial external fetch and three scheduled retries.
	defaultFetchMaxAttempts = 4
	// DefaultAdminNotifyDebounce bounds how often repeated failures emit unread admin notifications.
	DefaultAdminNotifyDebounce = 15 * time.Minute
)

// ErrInCooldown is returned when an enqueue is skipped because the resource is in cooldown.
var ErrInCooldown = errors.New("job target is in cooldown")

// EnqueueRequest describes a durable job to enqueue or coalesce.
type EnqueueRequest struct {
	JobType         string
	TargetType      string
	TargetID        int64
	TargetKey       string
	TargetRevision  int64
	MaxAttempts     int32
	CreatedByUserID sql.NullInt32
	ForceRetry      bool
}

// ClaimedJob pairs a claimed DurableJob row with its active fencing token.
type ClaimedJob struct {
	Job        *db.DurableJob
	LeaseOwner string
	LeaseToken string
}

// HandlerFunc executes one claimed durable job.
type HandlerFunc func(ctx context.Context, job *db.DurableJob) error

// Service coordinates durable job persistence, deduplication, leasing, cooldowns, and worker execution.
type Service struct {
	Queries                  db.Querier
	Now                      func() time.Time
	Wakeup                   chan struct{}
	CooldownDuration         time.Duration
	LeaseDuration            time.Duration
	AdminNotifyDebounce      time.Duration
	NotifyAdmins             func(ctx context.Context, message, link string) error
	recordAdminFailureAlerts bool
}

// NewService constructs a Service with sensible defaults.
func NewService(q db.Querier) *Service {
	return &Service{
		Queries:             q,
		Now:                 func() time.Time { return time.Now().UTC() },
		Wakeup:              make(chan struct{}, 1),
		CooldownDuration:    DefaultCooldown,
		LeaseDuration:       DefaultLeaseDuration,
		AdminNotifyDebounce: DefaultAdminNotifyDebounce,
	}
}

func (s *Service) now() time.Time {
	if s != nil && s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func (s *Service) cooldownDuration() time.Duration {
	if s != nil && s.CooldownDuration > 0 {
		return s.CooldownDuration
	}
	return DefaultCooldown
}

func (s *Service) leaseDuration() time.Duration {
	if s != nil && s.LeaseDuration > 0 {
		return s.LeaseDuration
	}
	return DefaultLeaseDuration
}

func (s *Service) notifyDebounce() time.Duration {
	if s != nil && s.AdminNotifyDebounce > 0 {
		return s.AdminNotifyDebounce
	}
	return DefaultAdminNotifyDebounce
}

// DedupeKey returns the canonical deduplication key for a job type and target key.
func DedupeKey(jobType, targetKey string) string {
	return strings.TrimSpace(jobType) + ":" + strings.TrimSpace(targetKey)
}

// SignalWakeup triggers a non-blocking wakeup for in-process worker loops.
func (s *Service) SignalWakeup() {
	if s == nil || s.Wakeup == nil {
		return
	}
	select {
	case s.Wakeup <- struct{}{}:
	default:
	}
}

// Enqueue creates or coalesces an active durable job for the given target.
// If an active job is already pending or running, its revision/rerun flag is bumped
// instead of creating a duplicate row.
func (s *Service) Enqueue(ctx context.Context, req EnqueueRequest) (bool, error) {
	if s == nil || s.Queries == nil {
		return false, nil
	}
	dedupeKey := DedupeKey(req.JobType, req.TargetKey)
	now := s.now()

	if req.ForceRetry {
		if err := s.Queries.SystemClearJobCooldownWindow(ctx, db.SystemClearJobCooldownWindowParams{
			CooldownUntil: now,
			DedupeKey:     dedupeKey,
		}); err != nil {
			return false, fmt.Errorf("clear cooldown: %w", err)
		}
	} else {
		cd, err := s.Queries.SystemGetJobCooldown(ctx, dedupeKey)
		if err == nil && cd.CooldownUntil.After(now) {
			return false, ErrInCooldown
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("check cooldown: %w", err)
		}
	}

	maxAttempts := req.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = DefaultMaxAttempts
		if req.JobType == JobTypeExternalLink || req.JobType == JobTypeImageCache {
			maxAttempts = defaultFetchMaxAttempts
		}
	}

	updated, err := s.Queries.SystemBumpActiveDurableJob(ctx, db.SystemBumpActiveDurableJobParams{
		TargetRevision:   req.TargetRevision,
		TargetRevision_2: req.TargetRevision,
		RunAfter:         now,
		RunAfter_2:       now,
		ActiveDedupeKey:  sql.NullString{String: dedupeKey, Valid: true},
	})
	if err != nil {
		return false, fmt.Errorf("coalesce active durable job: %w", err)
	}
	if updated > 0 {
		s.SignalWakeup()
		return false, nil
	}

	_, err = s.Queries.SystemInsertDurableJob(ctx, db.SystemInsertDurableJobParams{
		JobType:         req.JobType,
		DedupeKey:       dedupeKey,
		ActiveDedupeKey: sql.NullString{String: dedupeKey, Valid: true},
		TargetType:      req.TargetType,
		TargetID:        req.TargetID,
		TargetKey:       req.TargetKey,
		TargetRevision:  req.TargetRevision,
		MaxAttempts:     maxAttempts,
		RunAfter:        now,
		CreatedByUserID: req.CreatedByUserID,
	})
	if err != nil {
		// Handle concurrent insert race on uq_durable_jobs_active_dedupe by coalescing.
		bumped, bumpErr := s.Queries.SystemBumpActiveDurableJob(ctx, db.SystemBumpActiveDurableJobParams{
			TargetRevision:   req.TargetRevision,
			TargetRevision_2: req.TargetRevision,
			RunAfter:         now,
			RunAfter_2:       now,
			ActiveDedupeKey:  sql.NullString{String: dedupeKey, Valid: true},
		})
		if bumpErr == nil && bumped > 0 {
			s.SignalWakeup()
			return false, nil
		}
		return false, fmt.Errorf("insert durable job: %w", err)
	}
	s.SignalWakeup()
	return true, nil
}

// EnqueueSearchDirty bumps the dirty revision for a searchable resource and coalesces a search_index job.
func (s *Service) EnqueueSearchDirty(ctx context.Context, targetType string, targetID int32) error {
	if s == nil || s.Queries == nil || targetID <= 0 {
		return nil
	}
	if err := s.Queries.SystemBumpSearchDirtyTarget(ctx, db.SystemBumpSearchDirtyTargetParams{
		TargetType: targetType,
		TargetID:   targetID,
	}); err != nil {
		return fmt.Errorf("bump search dirty target: %w", err)
	}
	state, err := s.Queries.SystemGetSearchDirtyTarget(ctx, db.SystemGetSearchDirtyTargetParams{
		TargetType: targetType,
		TargetID:   targetID,
	})
	if err != nil {
		return fmt.Errorf("read search dirty target: %w", err)
	}
	_, err = s.Enqueue(ctx, EnqueueRequest{
		JobType:        JobTypeSearchIndex,
		TargetType:     targetType,
		TargetID:       int64(targetID),
		TargetKey:      fmt.Sprintf("%s:%d", targetType, targetID),
		TargetRevision: state.Revision,
		ForceRetry:     true,
	})
	return err
}

// ClaimNext claims the next eligible pending or expired-lease job, optionally filtered by jobTypes.
func (s *Service) ClaimNext(ctx context.Context, owner string, jobTypes ...string) (*ClaimedJob, error) {
	if s == nil || s.Queries == nil {
		return nil, nil
	}
	now := s.now()
	for attempt := 0; attempt < 3; attempt++ {
		var candidate *db.DurableJob
		var err error
		if len(jobTypes) > 0 {
			candidate, err = s.Queries.SystemGetNextClaimableDurableJobByTypes(ctx, db.SystemGetNextClaimableDurableJobByTypesParams{
				JobTypes:       jobTypes,
				RunAfter:       now,
				LeaseExpiresAt: sql.NullTime{Time: now, Valid: true},
			})
		} else {
			candidate, err = s.Queries.SystemGetNextClaimableDurableJob(ctx, db.SystemGetNextClaimableDurableJobParams{
				RunAfter:       now,
				LeaseExpiresAt: sql.NullTime{Time: now, Valid: true},
			})
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("select claimable durable job: %w", err)
		}

		token, err := newLeaseToken()
		if err != nil {
			return nil, fmt.Errorf("generate lease token: %w", err)
		}
		expiresAt := now.Add(s.leaseDuration())
		rows, err := s.Queries.SystemClaimDurableJobByID(ctx, db.SystemClaimDurableJobByIDParams{
			LeaseOwner:       sql.NullString{String: owner, Valid: true},
			LeaseToken:       sql.NullString{String: token, Valid: true},
			LeaseExpiresAt:   sql.NullTime{Time: expiresAt, Valid: true},
			ID:               candidate.ID,
			RunAfter:         now,
			LeaseExpiresAt_2: sql.NullTime{Time: now, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("claim durable job %d: %w", candidate.ID, err)
		}
		if rows == 0 {
			continue
		}
		claimed, err := s.Queries.SystemGetDurableJobByID(ctx, candidate.ID)
		if err != nil {
			return nil, fmt.Errorf("reload claimed job %d: %w", candidate.ID, err)
		}
		return &ClaimedJob{
			Job:        claimed,
			LeaseOwner: owner,
			LeaseToken: token,
		}, nil
	}
	return nil, nil
}

// CompleteFenced marks a claimed job completed (or re-arms pending if needs_rerun was set mid-flight)
// only while its lease is current and unexpired.
func (s *Service) CompleteFenced(ctx context.Context, claimed *ClaimedJob) (bool, error) {
	if s == nil || s.Queries == nil || claimed == nil || claimed.Job == nil {
		return false, nil
	}
	now := s.now()
	rows, err := s.Queries.SystemCompleteDurableJobFenced(ctx, db.SystemCompleteDurableJobFencedParams{
		RunAfter:    now,
		CompletedAt: sql.NullTime{Time: now, Valid: true},
		ID:          claimed.Job.ID,
		LeaseToken:  sql.NullString{String: claimed.LeaseToken, Valid: true},
		NowAt:       sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		return false, fmt.Errorf("complete durable job %d: %w", claimed.Job.ID, err)
	}
	return rows > 0, nil
}

// FailOrRetryFenced records an execution error while the claim is active. If attempts remain,
// it schedules durable exponential backoff; otherwise it transitions to failed, persists
// a per-target cooldown, and emits a debounced unread admin notification.
func (s *Service) FailOrRetryFenced(ctx context.Context, claimed *ClaimedJob, jobErr error) (bool, error) {
	if s == nil || s.Queries == nil || claimed == nil || claimed.Job == nil {
		return false, nil
	}
	now := s.now()
	errText := ""
	if jobErr != nil {
		errText = jobErr.Error()
	}

	permanent, retryAfter := fetchRetryPolicy(claimed.Job.JobType, jobErr, now)
	if !permanent && claimed.Job.Attempts < claimed.Job.MaxAttempts {
		backoff := max(retryBackoff(claimed.Job.JobType, claimed.Job.Attempts), retryAfter)
		rows, err := s.Queries.SystemRetryDurableJobFenced(ctx, db.SystemRetryDurableJobFencedParams{
			RunAfter:   now.Add(backoff),
			LastError:  sql.NullString{String: errText, Valid: errText != ""},
			ID:         claimed.Job.ID,
			LeaseToken: sql.NullString{String: claimed.LeaseToken, Valid: true},
			NowAt:      sql.NullTime{Time: now, Valid: true},
		})
		if err != nil {
			return false, fmt.Errorf("schedule retry for durable job %d: %w", claimed.Job.ID, err)
		}
		return rows > 0, nil
	}

	tx, ok := s.Queries.(db.JobTransactor)
	if !ok {
		return false, fmt.Errorf("durable job transactions unavailable")
	}
	var applied, notify bool
	err := tx.WithJobTransaction(ctx, func(q db.Querier) error {
		rows, err := q.SystemFailDurableJobFenced(ctx, db.SystemFailDurableJobFencedParams{
			LastError:   sql.NullString{String: errText, Valid: errText != ""},
			CompletedAt: sql.NullTime{Time: now, Valid: true},
			ID:          claimed.Job.ID,
			LeaseToken:  sql.NullString{String: claimed.LeaseToken, Valid: true},
			NowAt:       sql.NullTime{Time: now, Valid: true},
		})
		if err != nil || rows == 0 {
			return err
		}
		applied = true
		notify, err = s.recordTerminalCooldown(ctx, q, claimed.Job.DedupeKey, errText, now, permanent)
		if err != nil || !s.recordAdminFailureAlerts || s.NotifyAdmins != nil {
			return err
		}
		return s.recordAdminFailureAlertsForJob(ctx, q, claimed.Job, jobErr, now)
	})
	if err != nil {
		return false, fmt.Errorf("record terminal durable job %d: %w", claimed.Job.ID, err)
	}
	if notify {
		msg := fmt.Sprintf("Durable %s job #%d failed after %d attempts: %s", claimed.Job.JobType, claimed.Job.ID, claimed.Job.Attempts, errText)
		if err := s.NotifyAdmins(ctx, msg, "/admin/jobs?status=failed"); err != nil {
			log.Printf("notify admins of durable job failure: %v", err)
		}
	}
	return applied, nil
}

func (s *Service) recordTerminalCooldown(ctx context.Context, q db.Querier, key, errText string, now time.Time, permanent bool) (bool, error) {
	cooldownUntil := now.Add(s.cooldownDuration())
	if permanent {
		// An explicit operator retry can lift this durable suppression.
		cooldownUntil = time.Date(9999, time.December, 31, 0, 0, 0, 0, time.UTC)
	}
	existing, err := q.SystemGetJobCooldown(ctx, key)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	notify := s.NotifyAdmins != nil && (existing == nil || !existing.LastNotifiedAt.Valid || now.Sub(existing.LastNotifiedAt.Time) >= s.notifyDebounce())

	var notifiedAt sql.NullTime
	if notify {
		notifiedAt = sql.NullTime{Time: now, Valid: true}
	}
	err = q.SystemUpsertJobCooldown(ctx, db.SystemUpsertJobCooldownParams{
		DedupeKey:      key,
		CooldownUntil:  cooldownUntil,
		LastError:      sql.NullString{String: errText, Valid: errText != ""},
		LastNotifiedAt: notifiedAt,
	})
	return notify, err
}

// ReleaseOnShutdownFenced returns an interrupted job to pending without charging an attempt.
func (s *Service) ReleaseOnShutdownFenced(ctx context.Context, claimed *ClaimedJob) (bool, error) {
	if s == nil || s.Queries == nil || claimed == nil || claimed.Job == nil {
		return false, nil
	}
	rows, err := s.Queries.SystemReleaseDurableJobOnShutdownFenced(ctx, db.SystemReleaseDurableJobOnShutdownFencedParams{
		RunAfter:   s.now(),
		ID:         claimed.Job.ID,
		LeaseToken: sql.NullString{String: claimed.LeaseToken, Valid: true},
	})
	if err != nil {
		return false, fmt.Errorf("release durable job %d on shutdown: %w", claimed.Job.ID, err)
	}
	return rows > 0, nil
}

// AdminList returns durable jobs filtered by optional status and jobType.
func (s *Service) AdminList(ctx context.Context, statusFilter, jobTypeFilter string, limit, offset int) ([]*db.DurableJob, error) {
	if s == nil || s.Queries == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.Queries.AdminListDurableJobs(ctx, db.AdminListDurableJobsParams{
		Column1: statusFilter,
		Status:  statusFilter,
		Column3: jobTypeFilter,
		JobType: jobTypeFilter,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
}

// AdminRetry resets a terminal job to pending and clears any active cooldown for its dedupe key.
func (s *Service) AdminRetry(ctx context.Context, id int64) (bool, error) {
	retried, err := s.updateAdminJobState(ctx, id, StatusCancelled, StatusPending, func(q db.Querier) (int64, error) {
		rows, err := q.AdminRetryDurableJob(ctx, id)
		if err != nil {
			return 0, err
		}
		job, err := q.SystemGetDurableJobByID(ctx, id)
		if err != nil || rows == 0 {
			return rows, err
		}
		err = q.SystemClearJobCooldownWindow(ctx, db.SystemClearJobCooldownWindowParams{
			CooldownUntil: s.now(),
			DedupeKey:     job.DedupeKey,
		})
		return rows, err
	})
	if err != nil {
		return false, fmt.Errorf("retry durable job %d: %w", id, err)
	}
	if retried {
		s.SignalWakeup()
	}
	return retried, nil
}

// AdminCancel cancels a pending or running job and stops image recovery for its source.
func (s *Service) AdminCancel(ctx context.Context, id int64) (bool, error) {
	cancelled, err := s.updateAdminJobState(ctx, id, StatusPending, StatusCancelled, func(q db.Querier) (int64, error) {
		return q.AdminCancelDurableJob(ctx, id)
	})
	if err != nil {
		return false, fmt.Errorf("cancel durable job %d: %w", id, err)
	}
	return cancelled, nil
}

func (s *Service) updateAdminJobState(ctx context.Context, id int64, previousImageStatus, nextImageStatus string, update func(db.Querier) (int64, error)) (bool, error) {
	if s == nil || s.Queries == nil {
		return false, nil
	}
	tx, ok := s.Queries.(db.JobTransactor)
	if !ok {
		return false, fmt.Errorf("durable job transactions unavailable")
	}
	var rows int64
	err := tx.WithJobTransaction(ctx, func(q db.Querier) error {
		var err error
		rows, err = update(q)
		if err != nil || rows == 0 {
			return err
		}
		return q.AdminSetImageCacheStateForDurableJob(ctx, db.AdminSetImageCacheStateForDurableJobParams{
			JobID: id, PreviousStatus: previousImageStatus, NextStatus: nextImageStatus,
		})
	})
	return err == nil && rows > 0, err
}

// AdminPurge deletes terminal durable jobs matching statusFilter (or all terminal jobs when empty).
func (s *Service) AdminPurge(ctx context.Context, statusFilter string) (int64, error) {
	if s == nil || s.Queries == nil {
		return 0, nil
	}
	return s.Queries.AdminPurgeTerminalDurableJobs(ctx, db.AdminPurgeTerminalDurableJobsParams{
		Column1: statusFilter,
		Status:  statusFilter,
	})
}

// PruneTerminal removes terminal jobs beyond the age or per-kind history limit.
// Active jobs and cooldowns are retained.
func (s *Service) PruneTerminal(ctx context.Context, olderThan time.Duration, maxPerKind int) (int64, error) {
	if s == nil || s.Queries == nil {
		return 0, nil
	}
	if maxPerKind <= 0 {
		return 0, fmt.Errorf("terminal job history limit must be positive")
	}
	cutoff := s.now().Add(-olderThan)
	return s.Queries.SystemPruneTerminalDurableJobs(ctx, db.SystemPruneTerminalDurableJobsParams{
		Cutoff:     sql.NullTime{Time: cutoff, Valid: true},
		MaxPerKind: int64(maxPerKind),
	})
}

// RunWorkerLoop runs a bounded, shutdown-aware worker loop for the given job types.
// It uses one timer at a time (never one per job) and releases interrupted jobs on shutdown
// without charging an attempt.
func (s *Service) RunWorkerLoop(ctx context.Context, owner string, pollInterval time.Duration, jobTypes []string, handler HandlerFunc) {
	if s == nil || s.Queries == nil || handler == nil {
		return
	}
	if pollInterval <= 0 {
		pollInterval = DefaultPollInterval
	}
	timer := time.NewTimer(pollInterval)
	defer timer.Stop()

	for {
		for {
			if ctx.Err() != nil {
				return
			}
			claimed, err := s.ClaimNext(ctx, owner, jobTypes...)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("durable job claim (%s): %v", owner, err)
				break
			}
			if claimed == nil {
				break
			}
			s.executeClaimed(ctx, claimed, handler)
		}

		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(pollInterval)

		select {
		case <-ctx.Done():
			return
		case <-s.Wakeup:
		case <-timer.C:
		}
	}
}

func (s *Service) executeClaimed(ctx context.Context, claimed *ClaimedJob, handler HandlerFunc) {
	err := handler(ctx, claimed.Job)
	if errors.Is(err, ErrLeaseLost) {
		return
	}
	if err != nil {
		if ctx.Err() != nil || errors.Is(err, context.Canceled) {
			relCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer cancel()
			if _, relErr := s.ReleaseOnShutdownFenced(relCtx, claimed); relErr != nil {
				log.Printf("release durable job %d on shutdown: %v", claimed.Job.ID, relErr)
			}
			return
		}
		if _, failErr := s.FailOrRetryFenced(ctx, claimed, err); failErr != nil {
			log.Printf("fail/retry durable job %d: %v", claimed.Job.ID, failErr)
		}
		return
	}
	if ctx.Err() != nil {
		relCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if _, relErr := s.ReleaseOnShutdownFenced(relCtx, claimed); relErr != nil {
			log.Printf("release durable job %d on shutdown: %v", claimed.Job.ID, relErr)
		}
		return
	}
	if _, compErr := s.CompleteFenced(ctx, claimed); compErr != nil {
		log.Printf("complete durable job %d: %v", claimed.Job.ID, compErr)
	}
}

func retryBackoff(jobType string, attempts int32) time.Duration {
	if jobType == JobTypeExternalLink || jobType == JobTypeImageCache {
		return time.Duration(min(3, max(1, attempts))) * 10 * time.Second
	}
	if attempts < 1 {
		attempts = 1
	}
	shift := attempts - 1
	if shift > 6 {
		shift = 6
	}
	d := 5 * time.Second << shift
	if d > 5*time.Minute {
		return 5 * time.Minute
	}
	return d
}

func newLeaseToken() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
