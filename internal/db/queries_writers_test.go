package db

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestQueries_ListWriters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)

	rows := sqlmock.NewRows([]string{"username", "count"}).AddRow("bob", 2)
	mock.ExpectQuery(regexp.QuoteMeta(listWritersForViewer)).
		WithArgs(int32(1), int32(1), int32(1), sqlmock.AnyArg(), int32(5), int32(0)).
		WillReturnRows(rows)

	res, err := q.ListWriters(context.Background(), ListWritersParams{ViewerID: 1, Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("ListWriters: %v", err)
	}
	if len(res) != 1 || res[0].Username.String != "bob" || res[0].Count != 2 {
		t.Fatalf("unexpected result %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_SearchWriters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)

	rows := sqlmock.NewRows([]string{"username", "count"}).AddRow("bob", 2)
	mock.ExpectQuery(regexp.QuoteMeta(searchWritersForViewer)).
		WithArgs(int32(1), "%bob%", "%bob%", int32(1), int32(1), sqlmock.AnyArg(), int32(5), int32(0)).
		WillReturnRows(rows)

	res, err := q.SearchWriters(context.Background(), SearchWritersParams{ViewerID: 1, Query: "bob", Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("SearchWriters: %v", err)
	}
	if len(res) != 1 || res[0].Username.String != "bob" || res[0].Count != 2 {
		t.Fatalf("unexpected result %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
