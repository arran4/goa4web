package linker

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/internal/db"
)

func TestUserCanCreateLink_Allowed(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)
	mock.ExpectQuery("SELECT 1 FROM grants").
		WithArgs(sqlmock.AnyArg(), "linker", sqlmock.AnyArg(), "post", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	ok, err := UserCanCreateLink(context.Background(), q, sql.NullInt32{Int32: 1, Valid: true}, 2)
	if err != nil {
		t.Fatalf("UserCanCreateLink: %v", err)
	}
	if !ok {
		t.Errorf("expected allowed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUserCanCreateLink_Denied(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)
	mock.ExpectQuery("SELECT 1 FROM grants").
		WithArgs(sqlmock.AnyArg(), "linker", sqlmock.AnyArg(), "post", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	ok, err := UserCanCreateLink(context.Background(), q, sql.NullInt32{Int32: 1, Valid: true}, 2)
	if err != nil {
		t.Fatalf("UserCanCreateLink: %v", err)
	}
	if ok {
		t.Errorf("expected denied")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
