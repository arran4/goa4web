package goa4web

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type recordMail struct{ to, sub, body string }

func (r *recordMail) Send(ctx context.Context, to, subject, body string) error {
	r.to, r.sub, r.body = to, subject, body
	return nil
}

func TestNotifyChange(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)
	mock.ExpectExec("INSERT INTO pending_emails").WithArgs("a@b.com", sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	ctx := context.WithValue(context.Background(), ContextValues("queries"), q)
	rec := &recordMail{}
	if err := notifyChange(ctx, rec, "a@b.com", "http://host"); err != nil {
		t.Fatalf("notify error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestNotifyChangeErrors(t *testing.T) {
	rec := &recordMail{}
	if err := notifyChange(context.Background(), rec, "", "p"); err == nil {
		t.Fatal("expected error for empty email")
	}
}
