package main

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestThreadDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	q := New(db)
	mock.ExpectExec("DeleteForumThread").
		WithArgs(int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("RebuildForumTopicByIdMetaColumns").
		WithArgs(int32(2)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := ThreadDelete(context.Background(), q, 1, 2); err != nil {
		t.Fatalf("ThreadDelete: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}
