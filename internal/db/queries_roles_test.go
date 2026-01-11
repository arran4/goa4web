package db

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestQueries_AdminUpdateRole(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	mock.ExpectExec(regexp.QuoteMeta(adminUpdateRole)).
		WithArgs("name", true, false, true, sqlmock.AnyArg(), int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := q.AdminUpdateRole(context.Background(), AdminUpdateRoleParams{Name: "name", CanLogin: true, IsAdmin: false, PrivateLabels: true, PublicProfileAllowedAt: sql.NullTime{Time: time.Now(), Valid: true}, ID: 1}); err != nil {
		t.Fatalf("AdminUpdateRole: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
