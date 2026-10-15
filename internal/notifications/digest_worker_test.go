package notifications

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

func TestNotificationDigestWorker_ScheduleDigest(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	bus := eventbus.NewBus()
	// No queries needed for this test
	n := New(WithConfig(cfg), WithBus(bus))

	// Capture events using SyncPublish hook
	var capturedEvents []eventbus.Message
	bus.SyncPublish = func(msg eventbus.Message) {
		capturedEvents = append(capturedEvents, msg)
	}

	targetTime := time.Now().UTC()
	err := n.ScheduleDigest(context.Background(), targetTime)
	if err != nil {
		t.Fatalf("ScheduleDigest error: %v", err)
	}

	if len(capturedEvents) != 1 {
		t.Fatalf("expected 1 event, got %d", len(capturedEvents))
	}
	evt, ok := capturedEvents[0].(eventbus.DigestRunEvent)
	if !ok {
		t.Fatalf("expected DigestRunEvent")
	}
	if !evt.Time.Equal(targetTime) {
		t.Fatalf("expected time %v, got %v", targetTime, evt.Time)
	}
}

func TestNotificationDigestWorker_SendDigest(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = conn.Close() }()
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

	// The queued digest records the exact notification snapshot. It must not
	// mark the notification read until the email worker receives provider
	// acceptance.
	mock.ExpectExec("INSERT INTO digest_deliveries").
		WithArgs(userID, true).
		WillReturnResult(sqlmock.NewResult(25, 1))
	mock.ExpectExec("INSERT INTO digest_delivery_notifications").
		WithArgs(int32(25), int32(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO pending_emails").
		WithArgs(userID, sqlmock.AnyArg(), false, int32(25)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect UpdateLastWeeklyDigestSentAt (since we'll request Weekly)
	mock.ExpectExec("UPDATE preferences SET last_weekly_digest_sent_at").
		WithArgs(sqlmock.AnyArg(), userID).
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

func TestNotificationDigestWorker_PreviewDoesNotCreateDismissalState(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = conn.Close() }()

	q := db.New(conn)
	n := New(WithQueries(q), WithConfig(config.NewRuntimeConfig()))
	userID := int32(101)
	rows := sqlmock.NewRows([]string{"id", "users_idusers", "link", "message", "created_at", "read_at"}).
		AddRow(1, userID, "/link", "Test Notification", time.Now(), nil)
	mock.ExpectQuery("SELECT id, users_idusers").
		WithArgs(userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)
	mock.ExpectExec("INSERT INTO pending_emails").
		WithArgs(userID, sqlmock.AnyArg(), false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := n.SendDigestToUser(context.Background(), userID, "user@example.com", true, true, DigestDaily); err != nil {
		t.Fatalf("SendDigestToUser preview: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("preview changed digest state: %v", err)
	}
}

func TestNotificationDigestWorker_DisabledDismissalPersistsOptOut(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = conn.Close() }()

	q := db.New(conn)
	n := New(WithQueries(q), WithConfig(config.NewRuntimeConfig()))
	userID := int32(101)
	rows := sqlmock.NewRows([]string{"id", "users_idusers", "link", "message", "created_at", "read_at"}).
		AddRow(1, userID, "/link", "Test Notification", time.Now(), nil)
	mock.ExpectQuery("SELECT id, users_idusers").
		WithArgs(userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)
	mock.ExpectExec("INSERT INTO digest_deliveries").
		WithArgs(userID, false).
		WillReturnResult(sqlmock.NewResult(25, 1))
	mock.ExpectExec("INSERT INTO digest_delivery_notifications").
		WithArgs(int32(25), int32(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO pending_emails").
		WithArgs(userID, sqlmock.AnyArg(), false, int32(25)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE preferences SET last_digest_sent_at").
		WithArgs(sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := n.SendDigestToUser(context.Background(), userID, "user@example.com", false, false, DigestDaily); err != nil {
		t.Fatalf("SendDigestToUser with dismissal disabled: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("disabled dismissal wrote unexpected state: %v", err)
	}
}
