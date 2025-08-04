package db

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestQueries_SystemListAllUsers(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"idusers", "username", "email"}).
		AddRow(1, "bob", "bob@example.com")
	mock.ExpectQuery(regexp.QuoteMeta(systemListAllUsers)).
		WillReturnRows(rows)

	res, err := q.SystemListAllUsers(context.Background())
	if err != nil {
		t.Fatalf("SystemListAllUsers: %v", err)
	}
	if len(res) != 1 || res[0].Idusers != 1 || res[0].Username.String != "bob" || res[0].Email != "bob@example.com" {
		t.Fatalf("unexpected result %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
