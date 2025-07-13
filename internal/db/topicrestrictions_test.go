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
		ViewRoleID:             sql.NullInt32{Int32: 1, Valid: true},
		ReplyRoleID:            sql.NullInt32{Int32: 1, Valid: true},
		NewthreadRoleID:        sql.NullInt32{Int32: 1, Valid: true},
		SeeRoleID:              sql.NullInt32{Int32: 1, Valid: true},
		InviteRoleID:           sql.NullInt32{Int32: 1, Valid: true},
		ReadRoleID:             sql.NullInt32{Int32: 1, Valid: true},
		ModRoleID:              sql.NullInt32{Int32: 1, Valid: true},
		AdminRoleID:            sql.NullInt32{Int32: 1, Valid: true},
	}

	mock.ExpectExec("INSERT INTO topic_permissions").
		WithArgs(
			arg.ForumtopicIdforumtopic,
			arg.ViewRoleID,
			arg.ReplyRoleID,
			arg.NewthreadRoleID,
			arg.SeeRoleID,
			arg.InviteRoleID,
			arg.ReadRoleID,
			arg.ModRoleID,
			arg.AdminRoleID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if err := q.UpsertForumTopicRestrictions(context.Background(), arg); err != nil {
		t.Fatalf("insert: %v", err)
	}

	arg.ViewRoleID.Int32 = 2
	mock.ExpectExec("INSERT INTO topic_permissions").
		WithArgs(
			arg.ForumtopicIdforumtopic,
			arg.ViewRoleID,
			arg.ReplyRoleID,
			arg.NewthreadRoleID,
			arg.SeeRoleID,
			arg.InviteRoleID,
			arg.ReadRoleID,
			arg.ModRoleID,
			arg.AdminRoleID,
		).
		WillReturnResult(sqlmock.NewResult(1, 2))
	if err := q.UpsertForumTopicRestrictions(context.Background(), arg); err != nil {
		t.Fatalf("update: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
