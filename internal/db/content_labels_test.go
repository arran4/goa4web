package db

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAddAndListContentPublicLabels(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	mock.ExpectExec(regexp.QuoteMeta(addContentPublicLabel)).
		WithArgs("thread", int32(1), "foo").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := q.AddContentPublicLabel(context.Background(), AddContentPublicLabelParams{
		Item:   "thread",
		ItemID: 1,
		Label:  "foo",
	}); err != nil {
		t.Fatalf("AddContentPublicLabel: %v", err)
	}

	rows := sqlmock.NewRows([]string{"item", "item_id", "label"}).
		AddRow("thread", 1, "foo")
	mock.ExpectQuery(regexp.QuoteMeta(listContentPublicLabels)).
		WithArgs("thread", int32(1)).
		WillReturnRows(rows)

	res, err := q.ListContentPublicLabels(context.Background(), ListContentPublicLabelsParams{Item: "thread", ItemID: 1})
	if err != nil {
		t.Fatalf("ListContentPublicLabels: %v", err)
	}
	if len(res) != 1 || res[0].Label != "foo" {
		t.Fatalf("unexpected result %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAddAndListContentPrivateLabels(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := New(conn)

	mock.ExpectExec(regexp.QuoteMeta(addContentPrivateLabel)).
		WithArgs("thread", int32(1), int32(2), "bar", false).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := q.AddContentPrivateLabel(context.Background(), AddContentPrivateLabelParams{
		Item:   "thread",
		ItemID: 1,
		UserID: 2,
		Label:  "bar",
		Invert: false,
	}); err != nil {
		t.Fatalf("AddContentPrivateLabel: %v", err)
	}

	rows := sqlmock.NewRows([]string{"item", "item_id", "user_id", "label", "invert"}).
		AddRow("thread", 1, 2, "bar", false)
	mock.ExpectQuery(regexp.QuoteMeta(listContentPrivateLabels)).
		WithArgs("thread", int32(1), int32(2)).
		WillReturnRows(rows)

	res, err := q.ListContentPrivateLabels(context.Background(), ListContentPrivateLabelsParams{Item: "thread", ItemID: 1, UserID: 2})
	if err != nil {
		t.Fatalf("ListContentPrivateLabels: %v", err)
	}
	if len(res) != 1 || res[0].Label != "bar" || res[0].Invert {
		t.Fatalf("unexpected result %+v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
