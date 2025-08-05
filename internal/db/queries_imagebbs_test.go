package db

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestQueries_ListBoardsByParentIDForLister(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

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

func TestQueries_ListImagePostsByBoardForLister_GlobalGrant(t *testing.T) {
	if !strings.Contains(listImagePostsByBoardForLister, "g.item_id = i.imageboard_idimageboard OR g.item_id IS NULL") {
		t.Fatalf("global grant clause missing")
	}

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"idimagepost", "forumthread_id", "users_idusers", "imageboard_idimageboard", "posted", "description", "thumbnail", "fullimage", "file_size", "approved", "deleted_at", "last_index", "username", "comments"}).
		AddRow(1, 0, 1, 2, nil, nil, nil, nil, 0, true, nil, nil, "alice", 0)

	mock.ExpectQuery(regexp.QuoteMeta(listImagePostsByBoardForLister)).
		WithArgs(int32(1), int32(2), sql.NullInt32{Int32: 1, Valid: true}, int32(5), int32(0)).
		WillReturnRows(rows)

	res, err := q.ListImagePostsByBoardForLister(context.Background(), ListImagePostsByBoardForListerParams{ListerID: 1, BoardID: 2, ListerUserID: sql.NullInt32{Int32: 1, Valid: true}, Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("ListImagePostsByBoardForLister: %v", err)
	}
	if len(res) != 1 || res[0].Idimagepost != 1 {
		t.Fatalf("unexpected result %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
