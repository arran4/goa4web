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
)

func TestNotificationsQueries(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	mock.ExpectExec("INSERT INTO notifications").WithArgs(int32(1), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := q.SystemCreateNotification(context.Background(), db.SystemCreateNotificationParams{RecipientID: 1, Link: sql.NullString{String: "/x", Valid: true}, Message: sql.NullString{String: "hi", Valid: true}}); err != nil {
		t.Fatalf("insert: %v", err)
	}
	rows := sqlmock.NewRows([]string{"cnt"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").WillReturnRows(rows)
	if c, err := q.CountUnreadNotificationsForUser(context.Background(), 1); err != nil || c != 1 {
		t.Fatalf("count=%d err=%v", c, err)
	}
	mock.ExpectQuery("SELECT id, users_idusers").WillReturnRows(sqlmock.NewRows([]string{"id", "users_idusers", "link", "message", "created_at", "read_at"}).AddRow(1, 1, "/x", "hi", time.Now(), nil))
	if _, err := q.ListUnreadNotificationsForLister(context.Background(), db.ListUnreadNotificationsForListerParams{ListerID: 1, Limit: 10, Offset: 0}); err != nil {
		t.Fatalf("get: %v", err)
	}
	mock.ExpectExec("UPDATE notifications SET read_at").WithArgs(int32(1), int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := q.MarkNotificationReadForLister(context.Background(), db.MarkNotificationReadForListerParams{ID: 1, ListerID: 1}); err != nil {
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

func TestNotifierNotifyAdmins(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	cfg := config.NewRuntimeConfig()
	cfg.EmailEnabled = true
	cfg.AdminNotify = true
	cfg.AdminEmails = "a@test"
	cfg.EmailFrom = "from@example.com"
	cfg.NotificationsEnabled = true

	mock.ExpectQuery("SystemGetUserByEmail").
		WithArgs(sql.NullString{String: "a@test", Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "a@test", "a"))
	mock.ExpectExec("INSERT INTO pending_emails").WithArgs(sql.NullInt32{Int32: 1, Valid: true}, sqlmock.AnyArg(), false).WillReturnResult(sqlmock.NewResult(1, 1))
	rec := &dummyProvider{}
	n := New(WithQueries(q), WithEmailProvider(rec), WithConfig(cfg))
	n.NotifyAdmins(context.Background(), &EmailTemplates{}, EmailData{})
	if rec.to != "" {
		t.Fatalf("expected no direct mail got %s", rec.to)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestNotifierInitialization(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	n := New(WithConfig(cfg))
	if n.Queries != nil {
		t.Fatalf("expected nil Queries")
	}
	conn, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	n = New(WithQueries(q), WithConfig(cfg))
	if n.Queries != q {
		t.Fatalf("queries not set via option")
	}
}
