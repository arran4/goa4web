package forum

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/internal/db"
)

func TestUserCanCreateThread_Allowed(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := db.New(sqldb)
	mock.ExpectQuery("SELECT 1 FROM grants").
		WithArgs(sqlmock.AnyArg(), "forum", sqlmock.AnyArg(), "post", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	ok, err := UserCanCreateThread(context.Background(), q, 1, 2)
	if err != nil {
		t.Fatalf("UserCanCreateThread: %v", err)
	}
	if !ok {
		t.Errorf("expected allowed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUserCanCreateThread_Denied(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := db.New(sqldb)
	mock.ExpectQuery("SELECT 1 FROM grants").
		WithArgs(sqlmock.AnyArg(), "forum", sqlmock.AnyArg(), "post", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	ok, err := UserCanCreateThread(context.Background(), q, 1, 2)
	if err != nil {
		t.Fatalf("UserCanCreateThread: %v", err)
	}
	if ok {
		t.Errorf("expected denied")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
