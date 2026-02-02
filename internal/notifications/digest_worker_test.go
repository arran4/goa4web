package notifications

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

func TestNotificationDigestWorker_Scheduler(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	cfg := config.NewRuntimeConfig()
	bus := eventbus.NewBus()
	n := New(WithQueries(q), WithConfig(cfg), WithBus(bus))

	// Test case: Last run was 2 hours ago. Should run for T-1 and T-0.
	now := time.Now().UTC().Truncate(time.Hour)
	lastRun := now.Add(-2 * time.Hour)

	// Expect GetSchedulerState
	rows := sqlmock.NewRows([]string{"task_name", "last_run_at", "metadata"}).
		AddRow(SchedulerTaskName, lastRun, nil)
	mock.ExpectQuery("SELECT task_name, last_run_at, metadata FROM scheduler_state").
		WithArgs(SchedulerTaskName).
		WillReturnRows(rows)

	// Capture events using SyncPublish hook to avoid channel buffer issues
	var capturedEvents []eventbus.Message
	bus.SyncPublish = func(msg eventbus.Message) {
		capturedEvents = append(capturedEvents, msg)
	}

	// Expect UpsertSchedulerState
	mock.ExpectExec("INSERT INTO scheduler_state").
		WithArgs(SchedulerTaskName, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute
	n.processScheduler(context.Background())

	// Verify events
	count := 0
	for _, msg := range capturedEvents {
		if _, ok := msg.(eventbus.DigestRunEvent); ok {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected 2 digest run events, got %d", count)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expectations: %s", err)
	}
}

func TestNotificationDigestWorker_SendDigest(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	cfg := config.NewRuntimeConfig()
	n := New(WithQueries(q), WithConfig(cfg))

	userID := int32(101)
	emailAddr := "user@example.com"

	// Expect ListUnreadNotifications
	rows := sqlmock.NewRows([]string{"id", "users_idusers", "link", "message", "created_at", "read_at"}).
		AddRow(1, userID, "/link", "Test Notification", time.Now(), nil)
	mock.ExpectQuery("SELECT id, users_idusers").
		WithArgs(userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	// Expect InsertPendingEmail (sending email)
	mock.ExpectExec("INSERT INTO pending_emails").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect UpdateLastWeeklyDigestSentAt (since we'll request Weekly)
	mock.ExpectExec("UPDATE preferences SET last_weekly_digest_sent_at").
		WithArgs(sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect SetNotificationsReadForListerBatch (markRead=true)
	// sqlc expands slice to params. id=1 -> IN (?)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE notifications SET read_at = NOW() WHERE users_idusers = ? AND id IN (?)")).
		WithArgs(userID, int32(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute
	err = n.SendDigestToUser(context.Background(), userID, emailAddr, true, false, DigestWeekly)
	if err != nil {
		t.Fatalf("SendDigestToUser failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expectations: %s", err)
	}
}
