package goa4web

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type dummyProvider struct{ to string }

func (r *dummyProvider) Send(ctx context.Context, to, subj, body string) error {
	r.to = to
	return nil
}

func TestNotifyThreadSubscribers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)
	rows := sqlmock.NewRows([]string{
		"idcomments", "forumthread_idforumthread", "users_idusers", "language_idlanguage",
		"written", "text", "idusers", "email", "passwd", "passwd_algorithm", "username",
		"idpreferences", "language_idlanguage_2", "users_idusers_2", "emailforumupdates",
		"page_size",
	}).AddRow(1, 2, 2, 1, nil, "t", 2, "e", "p", "", "bob", 1, 1, 2, 1, 10)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idcomments, forumthread_idforumthread, c.users_idusers, c.language_idlanguage, written, text, idusers, email, passwd, passwd_algorithm, username, idpreferences, p.language_idlanguage, p.users_idusers, emailforumupdates, page_size\nFROM comments c, users u, preferences p\nWHERE c.forumthread_idforumthread=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=c.users_idusers AND u.idusers!=?\nGROUP BY u.idusers")).
		WithArgs(int32(2), int32(1)).
		WillReturnRows(rows)
	rec := &dummyProvider{}
	notifyThreadSubscribers(context.Background(), rec, q, 2, 1, "/p")
	if rec.to != "bob" {
		t.Fatalf("expected mail to bob got %s", rec.to)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
