package db

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestQueries_ListWritersForLister(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"username", "count"}).AddRow("bob", 2)
	mock.ExpectQuery(regexp.QuoteMeta(listWritersForLister)).
		WithArgs(int32(1), int32(1), int32(1), sqlmock.AnyArg(), int32(5), int32(0)).
		WillReturnRows(rows)

	res, err := q.ListWritersForLister(context.Background(), ListWritersForListerParams{ListerID: 1, UserID: sql.NullInt32{Int32: 1, Valid: true}, Limit: 5, Offset: 0})
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

func TestQueries_ListWritersSearchForLister(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"username", "count"}).AddRow("bob", 2)
	mock.ExpectQuery(regexp.QuoteMeta(listWritersSearchForLister)).
		WithArgs(int32(1), "%bob%", "%bob%", int32(1), int32(1), sqlmock.AnyArg(), int32(5), int32(0)).
		WillReturnRows(rows)

	res, err := q.ListWritersSearchForLister(context.Background(), ListWritersSearchForListerParams{ListerID: 1, Query: "%bob%", UserID: sql.NullInt32{Int32: 1, Valid: true}, Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("ListWritersSearchForLister: %v", err)
	}
	if len(res) != 1 || res[0].Username.String != "bob" || res[0].Count != 2 {
		t.Fatalf("unexpected result %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestQueries_GetWritingForListerByID_GlobalGrant(t *testing.T) {
	if !strings.Contains(getWritingForListerByID, "g.item_id = w.idwriting OR g.item_id IS NULL") {
		t.Fatalf("global grant clause missing")
	}

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	rows := sqlmock.NewRows([]string{"idwriting", "users_idusers", "forumthread_id", "language_id", "writing_category_id", "title", "published", "timezone", "writing", "abstract", "private", "deleted_at", "last_index", "WriterId", "WriterUsername"}).
		AddRow(1, 1, 0, 0, 0, nil, nil, nil, nil, nil, false, nil, nil, 1, "bob")

	mock.ExpectQuery(regexp.QuoteMeta(getWritingForListerByID)).
		WithArgs(int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(rows)

	res, err := q.GetWritingForListerByID(context.Background(), GetWritingForListerByIDParams{ListerID: 1, Idwriting: 1, ListerMatchID: sql.NullInt32{Int32: 1, Valid: true}})
	if err != nil {
		t.Fatalf("GetWritingForListerByID: %v", err)
	}
	if res.Idwriting != 1 {
		t.Fatalf("unexpected result %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
