package db

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestQueries_AdminListAllUserIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)

	rows := sqlmock.NewRows([]string{"idusers"}).AddRow(1).AddRow(2)
	mock.ExpectQuery(regexp.QuoteMeta(adminListAllUserIDs)).
		WillReturnRows(rows)

	res, err := q.AdminListAllUserIDs(context.Background())
	if err != nil {
		t.Fatalf("AdminListAllUserIDs: %v", err)
	}
	if len(res) != 2 || res[0] != 1 || res[1] != 2 {
		t.Fatalf("unexpected result %v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_AdminListAllUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)

	rows := sqlmock.NewRows([]string{"idusers", "username", "email"}).
		AddRow(1, "bob", "bob@example.com")
	mock.ExpectQuery(regexp.QuoteMeta(adminListAllUsers)).
		WillReturnRows(rows)

	res, err := q.AdminListAllUsers(context.Background())
	if err != nil {
		t.Fatalf("AdminListAllUsers: %v", err)
	}
	if len(res) != 1 || res[0].Idusers != 1 || res[0].Username.String != "bob" || res[0].Email != "bob@example.com" {
		t.Fatalf("unexpected result %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
