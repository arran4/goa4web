package db

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestQueries_AdminListAllUserIDs(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

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
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"idusers", "username"}).
		AddRow(1, "bob")
	mock.ExpectQuery(regexp.QuoteMeta(adminListAllUsers)).
		WillReturnRows(rows)

	res, err := q.AdminListAllUsers(context.Background())
	if err != nil {
		t.Fatalf("AdminListAllUsers: %v", err)
	}
	if len(res) != 1 || res[0].Idusers != 1 || res[0].Username.String != "bob" {
		t.Fatalf("unexpected result %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_SystemListAllUsers(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"idusers", "username", "admin", "created_at", "deleted_at"}).
		AddRow(1, "bob", false, time.Now(), nil)
	mock.ExpectQuery(regexp.QuoteMeta(systemListAllUsers)).
		WillReturnRows(rows)

	res, err := q.SystemListAllUsers(context.Background())
	if err != nil {
		t.Fatalf("SystemListAllUsers: %v", err)
	}
	if len(res) != 1 || res[0].Idusers != 1 || res[0].Username.String != "bob" {
		t.Fatalf("unexpected result %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_AdminDeleteUserByID(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	mock.ExpectExec(regexp.QuoteMeta(adminDeleteUserByID)).
		WithArgs(int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := q.AdminDeleteUserByID(context.Background(), 1); err != nil {
		t.Fatalf("AdminDeleteUserByID: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_AdminUpdateUsernameByID(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	mock.ExpectExec(regexp.QuoteMeta(adminUpdateUsernameByID)).
		WithArgs(sql.NullString{String: "bob", Valid: true}, int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := q.AdminUpdateUsernameByID(context.Background(), AdminUpdateUsernameByIDParams{Username: sql.NullString{String: "bob", Valid: true}, Idusers: 1}); err != nil {
		t.Fatalf("AdminUpdateUsernameByID: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
