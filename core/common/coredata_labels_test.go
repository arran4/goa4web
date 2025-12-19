package common_test

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestSetThreadPublicLabels(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 2

	rows := sqlmock.NewRows([]string{"item", "item_id", "label"}).
		AddRow("thread", 1, "foo").
		AddRow("thread", 1, "bar")
	mock.ExpectQuery("SELECT .* FROM content_public_labels").
		WithArgs("thread", int32(1)).
		WillReturnRows(rows)
	mock.ExpectQuery("SELECT .* FROM content_label_status").
		WithArgs("thread", int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"item", "item_id", "label"}))
	mock.ExpectExec("INSERT IGNORE INTO content_public_labels").
		WithArgs("thread", int32(1), "baz").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM content_public_labels").
		WithArgs("thread", int32(1), "foo").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := cd.SetThreadPublicLabels(1, []string{"bar", "baz"}); err != nil {
		t.Fatalf("SetThreadPublicLabels: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSetThreadPrivateLabels(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 2

	rows := sqlmock.NewRows([]string{"item", "item_id", "user_id", "label", "invert"}).
		AddRow("thread", 1, 2, "one", false).
		AddRow("thread", 1, 2, "two", false)
	mock.ExpectQuery("SELECT .* FROM content_private_labels").
		WithArgs("thread", int32(1), int32(2)).
		WillReturnRows(rows)
	mock.ExpectExec("INSERT IGNORE INTO content_private_labels").
		WithArgs("thread", int32(1), int32(2), "three", false).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM content_private_labels").
		WithArgs("thread", int32(1), int32(2), "one").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := cd.SetThreadPrivateLabels(1, []string{"two", "three"}); err != nil {
		t.Fatalf("SetThreadPrivateLabels: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPrivateLabelsDefaultAndInversion(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 2

	// Default case: no stored rows should return new and unread labels.
	mock.ExpectQuery("SELECT .* FROM content_private_labels").
		WithArgs("thread", int32(1), int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"item", "item_id", "user_id", "label", "invert"}))

	labels, err := cd.PrivateLabels("thread", 1)
	if err != nil {
		t.Fatalf("PrivateLabels default: %v", err)
	}
	expected := []string{"new", "unread"}
	if !reflect.DeepEqual(labels, expected) {
		t.Fatalf("default labels %+v, want %+v", labels, expected)
	}

	// Inversion case: storing an inverted new label removes it from the result.
	rows := sqlmock.NewRows([]string{"item", "item_id", "user_id", "label", "invert"}).
		AddRow("thread", 1, 2, "new", true).
		AddRow("thread", 1, 2, "foo", false)
	mock.ExpectQuery("SELECT .* FROM content_private_labels").
		WithArgs("thread", int32(1), int32(2)).
		WillReturnRows(rows)

	labels, err = cd.PrivateLabels("thread", 1)
	if err != nil {
		t.Fatalf("PrivateLabels invert: %v", err)
	}
	expected = []string{"unread", "foo"}
	if !reflect.DeepEqual(labels, expected) {
		t.Fatalf("inverted labels %+v, want %+v", labels, expected)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestClearThreadPrivateLabelStatus(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	cd := common.NewTestCoreData(t, q)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM content_private_labels")).
		WithArgs("thread", int32(1), "unread").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := cd.ClearThreadPrivateLabelStatus(1); err != nil {
		t.Fatalf("ClearThreadPrivateLabelStatus: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPrivateLabelsTopicExcludesStatus(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 2

	mock.ExpectQuery("SELECT .* FROM content_private_labels").
		WithArgs("topic", int32(1), int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"item", "item_id", "user_id", "label", "invert"}))

	labels, err := cd.PrivateLabels("topic", 1)
	if err != nil {
		t.Fatalf("PrivateLabels topic: %v", err)
	}
	if len(labels) != 0 {
		t.Fatalf("expected no labels for topic, got %+v", labels)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSetWritingPublicLabels(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 2

	mock.ExpectQuery("SELECT .* FROM content_public_labels").
		WithArgs("writing", int32(5)).
		WillReturnRows(sqlmock.NewRows([]string{"item", "item_id", "label"}))
	ownerRows := sqlmock.NewRows([]string{"item", "item_id", "label"}).
		AddRow("writing", 5, "a").
		AddRow("writing", 5, "b")
	mock.ExpectQuery("SELECT .* FROM content_label_status").
		WithArgs("writing", int32(5)).
		WillReturnRows(ownerRows)
	mock.ExpectExec("INSERT IGNORE INTO content_label_status").
		WithArgs("writing", int32(5), "c").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM content_label_status").
		WithArgs("writing", int32(5), "a").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := cd.SetWritingAuthorLabels(5, []string{"b", "c"}); err != nil {
		t.Fatalf("SetWritingAuthorLabels: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
