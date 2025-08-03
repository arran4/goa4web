package db

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"regexp"
)

func TestQueries_AdminUpdateRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)

	mock.ExpectExec(regexp.QuoteMeta(adminUpdateRole)).
		WithArgs("name", true, false, int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := q.AdminUpdateRole(context.Background(), AdminUpdateRoleParams{Name: "name", CanLogin: true, IsAdmin: false, ID: 1}); err != nil {
		t.Fatalf("AdminUpdateRole: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
