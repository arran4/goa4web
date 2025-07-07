package notifications

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/emailutil"
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

func (r *dummyProvider) Send(ctx context.Context, to, subj, textBody, htmlBody string) error {
	r.to = to
	return nil
}

func TestNotifyThreadSubscribers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	rows := sqlmock.NewRows([]string{
		"idcomments", "forumthread_idforumthread", "users_idusers", "language_idlanguage",
		"written", "text", "idusers", "email", "username",
		"idpreferences", "language_idlanguage_2", "users_idusers_2", "emailforumupdates",
		"page_size", "auto_subscribe_replies",
	}).AddRow(1, 2, 2, 1, nil, "t", 2, "e", "bob", 1, 1, 2, 1, 10, true)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT c.idcomments, c.forumthread_idforumthread, c.users_idusers, c.language_idlanguage, c.written, c.text, u.idusers, u.email, u.username, p.idpreferences, p.language_idlanguage, p.users_idusers, p.emailforumupdates, p.page_size, p.auto_subscribe_replies\nFROM comments c, users u, preferences p\nWHERE c.forumthread_idforumthread=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=c.users_idusers AND u.idusers!=?\nGROUP BY u.idusers")).
		WithArgs(int32(2), int32(1)).
		WillReturnRows(rows)
	rec := &dummyProvider{}
	emailutil.NotifyThreadSubscribers(context.Background(), rec, q, 2, 1, "/p")
	if rec.to != "bob" {
		t.Fatalf("expected mail to bob got %s", rec.to)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestNotifierNotifyAdmins(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	mock.ExpectQuery("SELECT u.email").
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("a@test"))
	mock.ExpectQuery("SELECT u.email").
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("a@test"))
	mock.ExpectQuery("UserByEmail").
		WithArgs(sql.NullString{String: "a@test", Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "a@test", "a"))
	mock.ExpectExec("INSERT INTO notifications").WithArgs(int32(1), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	rec := &dummyProvider{}
	n := Notifier{EmailProvider: rec, Queries: q}
	n.NotifyAdmins(context.Background(), "/p")
	if rec.to != "a@test" {
		t.Fatalf("expected mail to a@test got %s", rec.to)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
