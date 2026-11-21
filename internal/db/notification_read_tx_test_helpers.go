package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func runNotificationReadStateLinearizabilitySuite(t *testing.T, q Querier) {
	t.Helper()
	coord, ok := q.(NotificationReadStateCoordinator)
	if !ok {
		t.Fatalf("%T does not implement NotificationReadStateCoordinator", q)
	}

	t.Run("deliver_first_barrier_blocks_ack_until_commit", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		const (
			recipientID int32 = 101
			threadID    int32 = 201
			commentID   int32 = 301
			appendID    int64 = 1
		)
		pos := (int64(commentID) << 32) | appendID
		activityKey := fmt.Sprintf("reply:comment:%d:append:%d", commentID, appendID)

		deliverAtBarrier := make(chan struct{})
		ackStarted := make(chan struct{})
		ackFinished := make(chan error, 1)

		deliverCtx := WithReadStateTxHooks(ctx, ReadStateTxHooks{
			AfterDeliverMarkerRead: func(context.Context) {
				close(deliverAtBarrier)
				// Wait briefly until ack goroutine has started and is attempting to acquire
				// the lock; do not wait for ack to finish (it must block on our DB lock).
				select {
				case <-ackStarted:
				case <-ctx.Done():
				}
				// Give the ack goroutine a brief moment to hit the DB lock before we commit.
				time.Sleep(15 * time.Millisecond)
			},
		})

		deliverFinished := make(chan error, 1)
		go func() {
			deliverFinished <- coord.DeliverThreadNotificationForActivity(deliverCtx, SystemCreateNotificationForActivityParams{
				RecipientID:      recipientID,
				Link:             sql.NullString{String: "/thread/201#c301", Valid: true},
				Message:          sql.NullString{String: "append 1", Valid: true},
				ActivityKey:      sql.NullString{String: activityKey, Valid: true},
				Priority:         20,
				ResourceType:     sql.NullString{String: "thread", Valid: true},
				ResourceID:       sql.NullInt32{Int32: threadID, Valid: true},
				ActivityPosition: sql.NullInt64{Int64: pos, Valid: true},
			})
		}()

		<-deliverAtBarrier
		go func() {
			close(ackStarted)
			ackFinished <- coord.AcknowledgeThreadReadActivity(ctx, AcknowledgeThreadReadActivityParams{
				RecipientID:      recipientID,
				ThreadID:         threadID,
				CommentID:        commentID,
				ActivityPosition: pos,
			})
		}()

		if err := <-deliverFinished; err != nil {
			t.Fatalf("DeliverThreadNotificationForActivity: %v", err)
		}
		if err := <-ackFinished; err != nil {
			t.Fatalf("AcknowledgeThreadReadActivity: %v", err)
		}

		unreadCount, err := q.GetUnreadNotificationCountForLister(ctx, recipientID)
		if err != nil {
			t.Fatalf("GetUnreadNotificationCountForLister: %v", err)
		}
		if unreadCount != 0 {
			t.Fatalf("expected 0 unread notifications after racing deliver->ack, got %d", unreadCount)
		}
		marker, err := q.GetContentReadMarker(ctx, GetContentReadMarkerParams{
			Item:   "thread",
			ItemID: threadID,
			UserID: recipientID,
		})
		if err != nil {
			t.Fatalf("GetContentReadMarker: %v", err)
		}
		if marker.LastActivityPosition != pos {
			t.Fatalf("marker.LastActivityPosition = %d, want %d", marker.LastActivityPosition, pos)
		}
	})

	t.Run("ack_first_barrier_blocks_deliver_until_commit", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		const (
			recipientID int32 = 102
			threadID    int32 = 202
			commentID   int32 = 302
			appendID    int64 = 1
		)
		pos := (int64(commentID) << 32) | appendID
		activityKey := fmt.Sprintf("reply:comment:%d:append:%d", commentID, appendID)

		ackAtBarrier := make(chan struct{})
		deliverStarted := make(chan struct{})
		deliverFinished := make(chan error, 1)

		ackCtx := WithReadStateTxHooks(ctx, ReadStateTxHooks{
			AfterAckMarkerUpsert: func(context.Context) {
				close(ackAtBarrier)
				select {
				case <-deliverStarted:
				case <-ctx.Done():
				}
				time.Sleep(15 * time.Millisecond)
			},
		})

		ackFinished := make(chan error, 1)
		go func() {
			ackFinished <- coord.AcknowledgeThreadReadActivity(ackCtx, AcknowledgeThreadReadActivityParams{
				RecipientID:      recipientID,
				ThreadID:         threadID,
				CommentID:        commentID,
				ActivityPosition: pos,
			})
		}()

		<-ackAtBarrier
		go func() {
			close(deliverStarted)
			deliverFinished <- coord.DeliverThreadNotificationForActivity(ctx, SystemCreateNotificationForActivityParams{
				RecipientID:      recipientID,
				Link:             sql.NullString{String: "/thread/202#c302", Valid: true},
				Message:          sql.NullString{String: "append 1", Valid: true},
				ActivityKey:      sql.NullString{String: activityKey, Valid: true},
				Priority:         20,
				ResourceType:     sql.NullString{String: "thread", Valid: true},
				ResourceID:       sql.NullInt32{Int32: threadID, Valid: true},
				ActivityPosition: sql.NullInt64{Int64: pos, Valid: true},
			})
		}()

		if err := <-ackFinished; err != nil {
			t.Fatalf("AcknowledgeThreadReadActivity: %v", err)
		}
		if err := <-deliverFinished; err != nil {
			t.Fatalf("DeliverThreadNotificationForActivity: %v", err)
		}

		unreadCount, err := q.GetUnreadNotificationCountForLister(ctx, recipientID)
		if err != nil {
			t.Fatalf("GetUnreadNotificationCountForLister: %v", err)
		}
		if unreadCount != 0 {
			t.Fatalf("expected 0 unread notifications after racing ack->deliver, got %d", unreadCount)
		}
	})

	t.Run("consecutive_appends_and_multi_recipient_thread_isolation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		const (
			userA     int32 = 103
			userB     int32 = 104
			thread1   int32 = 203
			thread2   int32 = 204
			commentID int32 = 303
		)
		posAppend1 := (int64(commentID) << 32) | 1
		posAppend2 := (int64(commentID) << 32) | 2

		var wg sync.WaitGroup
		errs := make(chan error, 8)
		// Concurrently acknowledge append 1 for userA on thread1 while delivering:
		// - append 1 for userA on thread1 (must end READ)
		// - append 2 for userA on thread1 (must remain UNREAD)
		// - append 1 for userB on thread1 (different recipient -> must remain UNREAD)
		// - append 1 for userA on thread2 (different thread -> must remain UNREAD)
		wg.Add(5)
		go func() {
			defer wg.Done()
			errs <- coord.AcknowledgeThreadReadActivity(ctx, AcknowledgeThreadReadActivityParams{
				RecipientID:      userA,
				ThreadID:         thread1,
				CommentID:        commentID,
				ActivityPosition: posAppend1,
			})
		}()
		go func() {
			defer wg.Done()
			errs <- coord.DeliverThreadNotificationForActivity(ctx, SystemCreateNotificationForActivityParams{
				RecipientID:      userA,
				Link:             sql.NullString{String: "/thread/203#c303", Valid: true},
				Message:          sql.NullString{String: "uA t1 a1", Valid: true},
				ActivityKey:      sql.NullString{String: "reply:comment:303:append:1", Valid: true},
				Priority:         20,
				ResourceType:     sql.NullString{String: "thread", Valid: true},
				ResourceID:       sql.NullInt32{Int32: thread1, Valid: true},
				ActivityPosition: sql.NullInt64{Int64: posAppend1, Valid: true},
			})
		}()
		go func() {
			defer wg.Done()
			errs <- coord.DeliverThreadNotificationForActivity(ctx, SystemCreateNotificationForActivityParams{
				RecipientID:      userA,
				Link:             sql.NullString{String: "/thread/203#c303", Valid: true},
				Message:          sql.NullString{String: "uA t1 a2", Valid: true},
				ActivityKey:      sql.NullString{String: "reply:comment:303:append:2", Valid: true},
				Priority:         20,
				ResourceType:     sql.NullString{String: "thread", Valid: true},
				ResourceID:       sql.NullInt32{Int32: thread1, Valid: true},
				ActivityPosition: sql.NullInt64{Int64: posAppend2, Valid: true},
			})
		}()
		go func() {
			defer wg.Done()
			errs <- coord.DeliverThreadNotificationForActivity(ctx, SystemCreateNotificationForActivityParams{
				RecipientID:      userB,
				Link:             sql.NullString{String: "/thread/203#c303", Valid: true},
				Message:          sql.NullString{String: "uB t1 a1", Valid: true},
				ActivityKey:      sql.NullString{String: "reply:comment:303:append:1", Valid: true},
				Priority:         20,
				ResourceType:     sql.NullString{String: "thread", Valid: true},
				ResourceID:       sql.NullInt32{Int32: thread1, Valid: true},
				ActivityPosition: sql.NullInt64{Int64: posAppend1, Valid: true},
			})
		}()
		go func() {
			defer wg.Done()
			errs <- coord.DeliverThreadNotificationForActivity(ctx, SystemCreateNotificationForActivityParams{
				RecipientID:      userA,
				Link:             sql.NullString{String: "/thread/204#c303", Valid: true},
				Message:          sql.NullString{String: "uA t2 a1", Valid: true},
				ActivityKey:      sql.NullString{String: "reply:comment:304:append:1", Valid: true},
				Priority:         20,
				ResourceType:     sql.NullString{String: "thread", Valid: true},
				ResourceID:       sql.NullInt32{Int32: thread2, Valid: true},
				ActivityPosition: sql.NullInt64{Int64: posAppend1, Valid: true},
			})
		}()
		wg.Wait()
		close(errs)
		for err := range errs {
			if err != nil {
				t.Fatalf("concurrent read-state op failed: %v", err)
			}
		}

		unreadA, err := q.ListUnreadNotificationsForLister(ctx, ListUnreadNotificationsForListerParams{
			ListerID: userA,
			Limit:    20,
			Offset:   0,
		})
		if err != nil {
			t.Fatalf("list unread userA: %v", err)
		}
		if len(unreadA) != 2 {
			t.Fatalf("expected 2 unread notifications for userA (t1 append:2 and t2 append:1), got %d", len(unreadA))
		}
		for _, n := range unreadA {
			if n.ActivityKey.String == "reply:comment:303:append:1" {
				t.Fatalf("acknowledged notification reply:comment:303:append:1 for userA remained unread")
			}
		}

		unreadB, err := q.GetUnreadNotificationCountForLister(ctx, userB)
		if err != nil {
			t.Fatalf("count unread userB: %v", err)
		}
		if unreadB != 1 {
			t.Fatalf("expected 1 unread notification for independent userB, got %d", unreadB)
		}
	})
}

func TestRunReadStateTxWithRetry(t *testing.T) {
	var attempts atomic.Int32
	err := runReadStateTxWithRetry(context.Background(), func() error {
		if attempts.Add(1) < 3 {
			return errors.New("database is locked (SQLITE_BUSY)")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected transient busy error to recover on retry, got %v", err)
	}
	if attempts.Load() != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts.Load())
	}

	nonTransient := errors.New("constraint violation")
	attempts.Store(0)
	err = runReadStateTxWithRetry(context.Background(), func() error {
		attempts.Add(1)
		return nonTransient
	})
	if !errors.Is(err, nonTransient) {
		t.Fatalf("expected non-transient error immediately, got %v", err)
	}
	if attempts.Load() != 1 {
		t.Fatalf("expected 1 attempt for non-transient error, got %d", attempts.Load())
	}
}
