package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	// readStateTxMaxAttempts bounds retries on transient lock-contention or deadlock errors.
	readStateTxMaxAttempts = 6
	// readStateTxBaseBackoff is the initial retry delay for transient lock-contention errors.
	readStateTxBaseBackoff = 2 * time.Millisecond
)

// AcknowledgeThreadReadActivityParams describes an atomic thread read-marker update and
// matching notification acknowledgment for a single recipient.
type AcknowledgeThreadReadActivityParams struct {
	RecipientID      int32
	ThreadID         int32
	CommentID        int32
	ActivityPosition int64
}

// NotificationReadStateCoordinator defines linearizable, transactional operations for
// acknowledging thread read state and delivering activity-scoped notifications.
type NotificationReadStateCoordinator interface {
	AcknowledgeThreadReadActivity(ctx context.Context, arg AcknowledgeThreadReadActivityParams) error
	DeliverThreadNotificationForActivity(ctx context.Context, arg SystemCreateNotificationForActivityParams) error
}

// ReadStateTxHooks provides optional deterministic observation points inside read-state
// transactions for concurrency and interleaving verification.
type ReadStateTxHooks struct {
	AfterDeliverMarkerRead func(ctx context.Context)
	AfterAckMarkerUpsert   func(ctx context.Context)
}

type readStateTxHooksKey struct{}

// WithReadStateTxHooks attaches deterministic transaction hooks to ctx.
func WithReadStateTxHooks(ctx context.Context, hooks ReadStateTxHooks) context.Context {
	return context.WithValue(ctx, readStateTxHooksKey{}, hooks)
}

func readStateTxHooksFromContext(ctx context.Context) (ReadStateTxHooks, bool) {
	if ctx == nil {
		return ReadStateTxHooks{}, false
	}
	hooks, ok := ctx.Value(readStateTxHooksKey{}).(ReadStateTxHooks)
	return hooks, ok
}

func isTransientReadStateTxError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "database is locked") ||
		strings.Contains(msg, "sqlite_busy") ||
		strings.Contains(msg, "deadlock found") ||
		strings.Contains(msg, "lock wait timeout")
}

func runReadStateTxWithRetry(ctx context.Context, fn func() error) error {
	var err error
	for attempt := 0; attempt < readStateTxMaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = fn()
		if err == nil || !isTransientReadStateTxError(err) {
			return err
		}
		backoff := readStateTxBaseBackoff << attempt
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return err
}

func acknowledgeThreadReadActivityTx(
	ctx context.Context,
	beginTx func(context.Context, *sql.TxOptions) (*sql.Tx, error),
	withTx func(*sql.Tx) Querier,
	arg AcknowledgeThreadReadActivityParams,
) error {
	hooks, hasHooks := readStateTxHooksFromContext(ctx)
	return runReadStateTxWithRetry(ctx, func() error {
		tx, err := beginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin acknowledge read-state tx: %w", err)
		}
		defer func() {
			_ = tx.Rollback()
		}()

		qtx := withTx(tx)
		if err := qtx.UpsertContentReadMarker(ctx, UpsertContentReadMarkerParams{
			Item:                 "thread",
			ItemID:               arg.ThreadID,
			UserID:               arg.RecipientID,
			LastCommentID:        arg.CommentID,
			LastActivityPosition: arg.ActivityPosition,
		}); err != nil {
			return fmt.Errorf("upsert thread read marker: %w", err)
		}
		if hasHooks && hooks.AfterAckMarkerUpsert != nil {
			hooks.AfterAckMarkerUpsert(ctx)
		}
		if err := qtx.SystemAcknowledgeNotificationsForResource(ctx, SystemAcknowledgeNotificationsForResourceParams{
			RecipientID:      arg.RecipientID,
			ResourceType:     sql.NullString{String: "thread", Valid: true},
			ResourceID:       sql.NullInt32{Int32: arg.ThreadID, Valid: true},
			ActivityPosition: sql.NullInt64{Int64: arg.ActivityPosition, Valid: true},
		}); err != nil {
			return fmt.Errorf("acknowledge thread notifications: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit acknowledge read-state tx: %w", err)
		}
		return nil
	})
}

func deliverThreadNotificationForActivityTx(
	ctx context.Context,
	direct Querier,
	beginTx func(context.Context, *sql.TxOptions) (*sql.Tx, error),
	withTx func(*sql.Tx) Querier,
	arg SystemCreateNotificationForActivityParams,
) error {
	if !arg.ResourceType.Valid || arg.ResourceType.String != "thread" || !arg.ResourceID.Valid || !arg.ActivityPosition.Valid || arg.ReadAt.Valid {
		return direct.SystemCreateNotificationForActivity(ctx, arg)
	}
	hooks, hasHooks := readStateTxHooksFromContext(ctx)
	return runReadStateTxWithRetry(ctx, func() error {
		tx, err := beginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin notification delivery tx: %w", err)
		}
		defer func() {
			_ = tx.Rollback()
		}()

		qtx := withTx(tx)
		// Acquire the row/write lock on content_read_markers before reading the watermark so
		// concurrent acknowledgement and delivery serialize in the same lock order.
		if err := qtx.UpsertContentReadMarker(ctx, UpsertContentReadMarkerParams{
			Item:                 "thread",
			ItemID:               arg.ResourceID.Int32,
			UserID:               arg.RecipientID,
			LastCommentID:        0,
			LastActivityPosition: 0,
		}); err != nil {
			return fmt.Errorf("lock thread read marker: %w", err)
		}
		marker, err := qtx.GetContentReadMarkerForUpdate(ctx, GetContentReadMarkerForUpdateParams{
			Item:   "thread",
			ItemID: arg.ResourceID.Int32,
			UserID: arg.RecipientID,
		})
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("read locked thread marker: %w", err)
		}
		if hasHooks && hooks.AfterDeliverMarkerRead != nil {
			hooks.AfterDeliverMarkerRead(ctx)
		}

		deliverArg := arg
		if marker != nil {
			watermark := marker.LastActivityPosition
			if watermark == 0 && marker.LastCommentID > 0 {
				watermark = int64(marker.LastCommentID) << 32
			}
			if deliverArg.ActivityPosition.Int64 <= watermark {
				deliverArg.ReadAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
			}
		}
		if err := qtx.SystemCreateNotificationForActivity(ctx, deliverArg); err != nil {
			return fmt.Errorf("create notification for activity: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit notification delivery tx: %w", err)
		}
		return nil
	})
}

// AcknowledgeThreadReadActivity atomically updates the user's thread read watermark and marks
// matching notifications read in a single transaction.
func (q *Queries) AcknowledgeThreadReadActivity(ctx context.Context, arg AcknowledgeThreadReadActivityParams) error {
	return acknowledgeThreadReadActivityTx(ctx, q.BeginTx, func(tx *sql.Tx) Querier {
		return q.WithTx(tx)
	}, arg)
}

// DeliverThreadNotificationForActivity atomically evaluates the recipient's thread read watermark
// under lock and inserts or updates the activity notification in a single transaction.
func (q *Queries) DeliverThreadNotificationForActivity(ctx context.Context, arg SystemCreateNotificationForActivityParams) error {
	return deliverThreadNotificationForActivityTx(ctx, q, q.BeginTx, func(tx *sql.Tx) Querier {
		return q.WithTx(tx)
	}, arg)
}
