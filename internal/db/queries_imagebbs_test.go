package db

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestQueries_ListBoardsByParentIDForLister(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)

	rows := sqlmock.NewRows([]string{"idimageboard", "imageboard_idimageboard", "title", "description", "approval_required"}).AddRow(1, 0, nil, nil, 0)
	viewer := sql.NullInt32{}
	mock.ExpectQuery(regexp.QuoteMeta(listBoardsByParentIDForLister)).
		WithArgs(int32(1), int32(0), viewer, int32(5), int32(0)).
		WillReturnRows(rows)

	res, err := q.ListBoardsByParentIDForLister(context.Background(), ListBoardsByParentIDForListerParams{ListerID: 1, ParentID: 0, ListerUserID: viewer, Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("ListBoardsByParentIDForLister: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("unexpected result count %d", len(res))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_GetImageBoardByIDForLister(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)

	rows := sqlmock.NewRows([]string{"idimageboard", "imageboard_idimageboard", "title", "description", "approval_required"}).
		AddRow(2, 0, nil, nil, 0)
	viewer := sql.NullInt32{}
	mock.ExpectQuery(regexp.QuoteMeta(getImageBoardByIDForLister)).
		WithArgs(int32(1), int32(2), viewer).
		WillReturnRows(rows)

	res, err := q.GetImageBoardByIDForLister(context.Background(), GetImageBoardByIDForListerParams{ListerID: 1, BoardID: 2, ListerUserID: viewer})
	if err != nil {
		t.Fatalf("GetImageBoardByIDForLister: %v", err)
	}
	if res.Idimageboard != 2 {
		t.Fatalf("unexpected board id %d", res.Idimageboard)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
