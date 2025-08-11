package db

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAddAndListTopicPublicLabels(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	mock.ExpectExec(regexp.QuoteMeta(addTopicPublicLabel)).
		WithArgs(int32(1), "foo").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := q.AddTopicPublicLabel(context.Background(), AddTopicPublicLabelParams{
		ForumtopicIdforumtopic: 1,
		Label:                  "foo",
	}); err != nil {
		t.Fatalf("AddTopicPublicLabel: %v", err)
	}

	rows := sqlmock.NewRows([]string{"forumtopic_idforumtopic", "label"}).
		AddRow(1, "foo")
	mock.ExpectQuery(regexp.QuoteMeta(listTopicPublicLabels)).
		WithArgs(int32(1)).
		WillReturnRows(rows)

	res, err := q.ListTopicPublicLabels(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListTopicPublicLabels: %v", err)
	}
	if len(res) != 1 || res[0].Label != "foo" {
		t.Fatalf("unexpected result %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAddAndListTopicPrivateLabels(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	mock.ExpectExec(regexp.QuoteMeta(addTopicPrivateLabel)).
		WithArgs(int32(1), int32(2), "bar", false).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := q.AddTopicPrivateLabel(context.Background(), AddTopicPrivateLabelParams{
		ForumtopicIdforumtopic: 1,
		UsersIdusers:           2,
		Label:                  "bar",
		Invert:                 false,
	}); err != nil {
		t.Fatalf("AddTopicPrivateLabel: %v", err)
	}

	rows := sqlmock.NewRows([]string{"forumtopic_idforumtopic", "users_idusers", "label", "invert"}).
		AddRow(1, 2, "bar", false)
	mock.ExpectQuery(regexp.QuoteMeta(listTopicPrivateLabels)).
		WithArgs(int32(1), int32(2)).
		WillReturnRows(rows)

	res, err := q.ListTopicPrivateLabels(context.Background(), ListTopicPrivateLabelsParams{ForumtopicIdforumtopic: 1, UsersIdusers: 2})
	if err != nil {
		t.Fatalf("ListTopicPrivateLabels: %v", err)
	}
	if len(res) != 1 || res[0].Label != "bar" || res[0].Invert {
		t.Fatalf("unexpected result %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
