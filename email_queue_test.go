package goa4web

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type emailRecordProvider struct {
	to   string
	subj string
	body string
}

func (r *emailRecordProvider) Send(ctx context.Context, to, sub, body string) error {
	r.to = to
	r.subj = sub
	r.body = body
	return nil
}

func TestInsertPendingEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	q := New(db)
	mock.ExpectExec("INSERT INTO pending_emails").WithArgs("t@test", "sub", "body").WillReturnResult(sqlmock.NewResult(1, 1))

	if err := q.InsertPendingEmail(context.Background(), InsertPendingEmailParams{ToEmail: "t@test", Subject: "sub", Body: "body"}); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestEmailQueueWorker(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)
	rows := sqlmock.NewRows([]string{"id", "to_email", "subject", "body"}).AddRow(1, "a@test", "s", "b")
	mock.ExpectQuery("SELECT id, to_email").WillReturnRows(rows)
	mock.ExpectExec("UPDATE pending_emails SET sent_at").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))

	rec := &emailRecordProvider{}
	processPendingEmail(context.Background(), q, rec)

	if rec.to != "a@test" {
		t.Fatalf("got %q", rec.to)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
