package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUpsertForumTopicRestrictionsUpdates(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := New(sqldb)

	arg := UpsertForumTopicRestrictionsParams{
		ForumtopicIdforumtopic: 1,
		Viewlevel:              sql.NullInt32{Int32: 1, Valid: true},
		Replylevel:             sql.NullInt32{Int32: 1, Valid: true},
		Newthreadlevel:         sql.NullInt32{Int32: 1, Valid: true},
		Seelevel:               sql.NullInt32{Int32: 1, Valid: true},
		Invitelevel:            sql.NullInt32{Int32: 1, Valid: true},
		Readlevel:              sql.NullInt32{Int32: 1, Valid: true},
		Modlevel:               sql.NullInt32{Int32: 1, Valid: true},
		Adminlevel:             sql.NullInt32{Int32: 1, Valid: true},
	}

	mock.ExpectExec("INSERT INTO topicrestrictions").
		WithArgs(
			arg.ForumtopicIdforumtopic,
			arg.Viewlevel,
			arg.Replylevel,
			arg.Newthreadlevel,
			arg.Seelevel,
			arg.Invitelevel,
			arg.Readlevel,
			arg.Modlevel,
			arg.Adminlevel,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if err := q.UpsertForumTopicRestrictions(context.Background(), arg); err != nil {
		t.Fatalf("insert: %v", err)
	}

	arg.Viewlevel.Int32 = 2
	mock.ExpectExec("INSERT INTO topicrestrictions").
		WithArgs(
			arg.ForumtopicIdforumtopic,
			arg.Viewlevel,
			arg.Replylevel,
			arg.Newthreadlevel,
			arg.Seelevel,
			arg.Invitelevel,
			arg.Readlevel,
			arg.Modlevel,
			arg.Adminlevel,
		).
		WillReturnResult(sqlmock.NewResult(1, 2))
	if err := q.UpsertForumTopicRestrictions(context.Background(), arg); err != nil {
		t.Fatalf("update: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
