package main

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListUnsentPendingEmails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)
	rows := sqlmock.NewRows([]string{"id", "to_email", "subject", "body", "created_at"}).AddRow(1, "a@test", "s", "b", time.Now())
	mock.ExpectQuery("SELECT id, to_email, subject, body, created_at FROM pending_emails WHERE sent_at IS NULL ORDER BY id").WillReturnRows(rows)
	if _, err := q.ListUnsentPendingEmails(context.Background()); err != nil {
		t.Fatalf("list: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestRecentNotifications(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)
	rows := sqlmock.NewRows([]string{"id", "users_idusers", "link", "message", "created_at", "read_at"}).AddRow(1, 1, "/l", "m", time.Now(), nil)
	mock.ExpectQuery("SELECT id, users_idusers, link, message, created_at, read_at FROM notifications ORDER BY id DESC LIMIT ?").WithArgs(int32(5)).WillReturnRows(rows)
	if _, err := q.RecentNotifications(context.Background(), 5); err != nil {
		t.Fatalf("recent: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}
