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
