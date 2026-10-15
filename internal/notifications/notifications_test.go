package notifications

import (
	"context"
	"database/sql"
	"net/mail"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestNotificationsQueries(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = conn.Close() }()
	q := db.New(conn)
	mock.ExpectExec("INSERT INTO notifications").WithArgs(int32(1), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := q.SystemCreateNotification(context.Background(), db.SystemCreateNotificationParams{RecipientID: 1, Link: sql.NullString{String: "/x", Valid: true}, Message: sql.NullString{String: "hi", Valid: true}}); err != nil {
		t.Fatalf("insert: %v", err)
	}
	rows := sqlmock.NewRows([]string{"cnt"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").WillReturnRows(rows)
	if c, err := q.GetUnreadNotificationCountForLister(context.Background(), 1); err != nil || c != 1 {
		t.Fatalf("count=%d err=%v", c, err)
	}
	mock.ExpectQuery("SELECT notifications.id, notifications.users_idusers").WillReturnRows(sqlmock.NewRows([]string{"id", "users_idusers", "link", "message", "activity_key", "priority", "created_at", "read_at"}).AddRow(1, 1, "/x", "hi", nil, 0, time.Now(), nil))
	if _, err := q.ListUnreadNotificationsForLister(context.Background(), db.ListUnreadNotificationsForListerParams{ListerID: 1, Limit: 10, Offset: 0}); err != nil {
		t.Fatalf("get: %v", err)
	}
	mock.ExpectExec("UPDATE notifications SET read_at").WithArgs(int32(1), int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := q.SetNotificationReadForLister(context.Background(), db.SetNotificationReadForListerParams{ID: 1, ListerID: 1}); err != nil {
		t.Fatalf("mark: %v", err)
	}
	mock.ExpectExec("DELETE FROM notifications").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := q.AdminPurgeReadNotifications(context.Background()); err != nil {
		t.Fatalf("purge: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

type dummyProvider struct{ to string }

func (r *dummyProvider) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	r.to = to.Address
	return nil
}

func (r *dummyProvider) TestConfig(ctx context.Context) error { return nil }

func TestNotifierNotifyAdmins(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.SystemGetUserByEmailRow = &db.SystemGetUserByEmailRow{Idusers: 1, Email: "a@test", Username: sql.NullString{String: "a", Valid: true}}
	q.ListSubscribersForPatternsReturn = map[string][]int32{
		"notify:/admin/*": {1},
	}
	cfg := config.NewRuntimeConfig()
	cfg.EmailEnabled = true
	cfg.AdminNotify = true
	cfg.AdminEmails = "a@test"
	cfg.EmailFrom = "from@example.com"
	cfg.NotificationsEnabled = true

	rec := &dummyProvider{}
	n := New(WithQueries(q), WithEmailProvider(rec), WithConfig(cfg))
	_ = n.NotifyAdmins(context.Background(), &EmailTemplates{}, EmailData{})
	if rec.to != "" {
		t.Fatalf("expected no direct mail got %s", rec.to)
	}
	if len(q.SystemGetUserByEmailCalls) != 1 || q.SystemGetUserByEmailCalls[0] != "a@test" {
		t.Fatalf("expected user lookup for a@test, got %v", q.SystemGetUserByEmailCalls)
	}
	if len(q.InsertPendingEmailCalls) != 1 {
		t.Fatalf("expected 1 queued email got %d", len(q.InsertPendingEmailCalls))
	}
	if !q.InsertPendingEmailCalls[0].ToUserID.Valid || q.InsertPendingEmailCalls[0].ToUserID.Int32 != 1 {
		t.Fatalf("expected queued email for user 1, got %+v", q.InsertPendingEmailCalls[0].ToUserID)
	}
	if q.InsertPendingEmailCalls[0].DirectEmail {
		t.Fatalf("expected queued email to be non-direct")
	}
	if q.InsertPendingEmailCalls[0].Body == "" {
		t.Fatalf("expected non-empty email body")
	}
}

func TestNotifierInitialization(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	n := New(WithConfig(cfg))
	if n.Queries != nil {
		t.Fatalf("expected nil Queries")
	}
	q := testhelpers.NewQuerierStub()
	n = New(WithQueries(q), WithConfig(cfg))
	if n.Queries != q {
		t.Fatalf("queries not set via option")
	}
}

func TestSendInternalNotificationForEventUsesStableActivityKey(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	n := New(WithQueries(q), WithConfig(config.NewRuntimeConfig()))
	evt := eventbus.TaskEvent{Task: tasks.TaskString("Reply"), ActivityID: "comment:42"}

	if err := n.sendInternalNotificationForEvent(context.Background(), evt, 7, "/forum/topic/1/thread/2#c42", "subscriber message", notificationPrioritySubscriber); err != nil {
		t.Fatalf("send subscriber contribution: %v", err)
	}
	if err := n.sendInternalNotificationForEvent(context.Background(), evt, 7, "", "admin message", notificationPriorityAdmin); err != nil {
		t.Fatalf("send admin contribution: %v", err)
	}

	if len(q.SystemCreateNotificationForActivityCalls) != 2 {
		t.Fatalf("activity upserts = %d, want 2", len(q.SystemCreateNotificationForActivityCalls))
	}
	first := q.SystemCreateNotificationForActivityCalls[0]
	if first.ActivityKey.String != "reply:comment:42" || first.Priority != notificationPrioritySubscriber {
		t.Errorf("subscriber activity upsert = %#v", first)
	}
	second := q.SystemCreateNotificationForActivityCalls[1]
	if second.Priority != notificationPriorityAdmin {
		t.Errorf("admin priority = %d, want %d", second.Priority, notificationPriorityAdmin)
	}
}
