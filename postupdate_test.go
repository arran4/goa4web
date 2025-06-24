package goa4web

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	q := New(db)
	mock.ExpectExec("RecalculateForumThreadByIdMetaData").
		WithArgs(int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("RebuildForumTopicByIdMetaColumns").
		WithArgs(int32(2)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := PostUpdate(context.Background(), q, 1, 2); err != nil {
		t.Fatalf("PostUpdate: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}
