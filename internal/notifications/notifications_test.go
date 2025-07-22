package notifications

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"net/mail"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

func TestNotificationsQueries(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	mock.ExpectExec("INSERT INTO notifications").WithArgs(int32(1), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := q.InsertNotification(context.Background(), dbpkg.InsertNotificationParams{UsersIdusers: 1, Link: sql.NullString{String: "/x", Valid: true}, Message: sql.NullString{String: "hi", Valid: true}}); err != nil {
		t.Fatalf("insert: %v", err)
	}
	rows := sqlmock.NewRows([]string{"cnt"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").WillReturnRows(rows)
	if c, err := q.CountUnreadNotifications(context.Background(), 1); err != nil || c != 1 {
		t.Fatalf("count=%d err=%v", c, err)
	}
	mock.ExpectQuery("SELECT id, users_idusers").WillReturnRows(sqlmock.NewRows([]string{"id", "users_idusers", "link", "message", "created_at", "read_at"}).AddRow(1, 1, "/x", "hi", time.Now(), nil))
	if _, err := q.GetUnreadNotifications(context.Background(), 1); err != nil {
		t.Fatalf("get: %v", err)
	}
	mock.ExpectExec("UPDATE notifications SET read_at").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := q.MarkNotificationRead(context.Background(), 1); err != nil {
		t.Fatalf("mark: %v", err)
	}
	mock.ExpectExec("DELETE FROM notifications").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := q.PurgeReadNotifications(context.Background()); err != nil {
		t.Fatalf("purge: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestNotificationsFeed(t *testing.T) {
	r := httptest.NewRequest("GET", "/notifications/rss", nil)
	n := []*dbpkg.Notification{{ID: 1, Link: sql.NullString{String: "/l", Valid: true}, Message: sql.NullString{String: "m", Valid: true}}}
	feed := NotificationsFeed(r, n)
	if len(feed.Items) != 1 || feed.Items[0].Link.Href != "/l" {
		t.Fatalf("feed item incorrect")
	}
}

type dummyProvider struct{ to string }

func (r *dummyProvider) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	r.to = to.Address
	return nil
}

func TestNotifierNotifyAdmins(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	origCfg := config.AppRuntimeConfig
	config.AppRuntimeConfig.EmailEnabled = true
	config.AppRuntimeConfig.AdminNotify = true
	config.AppRuntimeConfig.AdminEmails = "a@test"
	config.AppRuntimeConfig.EmailFrom = "from@example.com"
	config.AppRuntimeConfig.NotificationsEnabled = true
	t.Cleanup(func() { config.AppRuntimeConfig = origCfg })
	mock.ExpectQuery("UserByEmail").
		WithArgs(sql.NullString{String: "a@test", Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "a@test", "a"))
	mock.ExpectExec("INSERT INTO pending_emails").WithArgs(sql.NullInt32{Int32: 1, Valid: true}, sqlmock.AnyArg(), false).WillReturnResult(sqlmock.NewResult(1, 1))
	rec := &dummyProvider{}
	n := New(WithQueries(q), WithEmailProvider(rec))
	n.NotifyAdmins(context.Background(), &EmailTemplates{}, EmailData{})
	if rec.to != "" {
		t.Fatalf("expected no direct mail got %s", rec.to)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
