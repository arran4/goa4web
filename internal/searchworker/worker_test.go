package searchworker

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

func TestProcessIndex(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	q := dbpkg.New(db)
	mock.ExpectExec("CreateSearchWord").
		WithArgs("hello").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("AddToForumCommentSearch").
		WithArgs(int32(5), int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	req := IndexRequest{Type: IndexForum, ID: 5, Text: "hello"}
	if err := processIndex(context.Background(), q, req); err != nil {
		t.Fatalf("processIndex: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}
