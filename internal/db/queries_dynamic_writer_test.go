package db

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestListWriters(t *testing.T) {
	dbconn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer dbconn.Close()

	q := New(dbconn)
	rows := sqlmock.NewRows([]string{"username", "count"}).
		AddRow(sql.NullString{String: "alice", Valid: true}, int64(3))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.username, COUNT(w.idwriting) AS count")).
		WithArgs(int32(1), int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}, int32(5), int32(0)).
		WillReturnRows(rows)

	res, err := q.ListWriters(context.Background(), ListWritersParams{ViewerID: 1, Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("ListWriters: %v", err)
	}
	if len(res) != 1 || res[0].Count != 3 || !res[0].Username.Valid || res[0].Username.String != "alice" {
		t.Fatalf("unexpected result %+v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSearchWriters(t *testing.T) {
	dbconn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer dbconn.Close()

	q := New(dbconn)
	rows := sqlmock.NewRows([]string{"username", "count"}).
		AddRow(sql.NullString{String: "bob", Valid: true}, int64(2))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.username, COUNT(w.idwriting) AS count")).
		WithArgs(int32(1), "%query%", "%query%", int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}, int32(5), int32(0)).
		WillReturnRows(rows)

	res, err := q.SearchWriters(context.Background(), SearchWritersParams{ViewerID: 1, Query: "query", Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("SearchWriters: %v", err)
	}
	if len(res) != 1 || res[0].Count != 2 || !res[0].Username.Valid || res[0].Username.String != "bob" {
		t.Fatalf("unexpected result %+v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
