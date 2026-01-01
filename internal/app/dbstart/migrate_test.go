package dbstart

import (
	"context"
	"database/sql"
	"testing"
	"testing/fstest"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestApply(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	mfs := fstest.MapFS{
		"0002.mysql.sql": {Data: []byte("CREATE TABLE t (id int);")},
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_version").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT version FROM schema_version").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("INSERT INTO schema_version").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE t").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("UPDATE schema_version SET version = ?").WithArgs(2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := Apply(context.Background(), conn, mfs, false, "mysql"); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestApplyWithDescription(t *testing.T) {
	t.Skip("Not supported atm - disabled")
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	mfs := fstest.MapFS{
		"0002_description.mysql.sql": {Data: []byte("CREATE TABLE t (id int);")},
	}

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_version").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT version FROM schema_version").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("INSERT INTO schema_version").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE t").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("UPDATE schema_version SET version = ?").WithArgs(2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := Apply(context.Background(), conn, mfs, false, "mysql"); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
