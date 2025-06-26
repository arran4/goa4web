package db

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDeactivateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)

	mock.ExpectExec("INSERT INTO deactivated_users").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE users").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO deactivated_comments").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE comments").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM permissions").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM sessions").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))

	if err := q.DeactivateUser(context.Background(), 1); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestRestoreUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)

	mock.ExpectExec("UPDATE users").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM deactivated_users").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE comments").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM deactivated_comments").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))

	if err := q.RestoreUser(context.Background(), 1); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}
