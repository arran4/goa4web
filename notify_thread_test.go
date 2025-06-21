package main

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type notifyRecordProvider struct{ to string }

func (r *notifyRecordProvider) Send(ctx context.Context, to, subj, body string) error {
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
		"written", "text", "idusers", "email", "passwd", "username",
		"idpreferences", "language_idlanguage_2", "users_idusers_2", "emailforumupdates", "page_size",
	}).AddRow(1, 2, 2, 1, nil, "t", 2, "e", "p", "bob", 1, 1, 2, 1, 15)
	mock.ExpectQuery(regexp.QuoteMeta(listUsersSubscribedToThread)).
		WithArgs(int32(2), int32(1)).
		WillReturnRows(rows)
	rec := &notifyRecordProvider{}
	notifyThreadSubscribers(context.Background(), rec, q, 2, 1, "/p")
	if rec.to != "bob" {
		t.Fatalf("expected mail to bob got %s", rec.to)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
