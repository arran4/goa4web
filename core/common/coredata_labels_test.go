package common_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/arran4/goa4web/config"
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
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
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
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
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

func TestSetWritingPublicLabels(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 2

	rows := sqlmock.NewRows([]string{"item", "item_id", "label"}).
		AddRow("writing", 5, "a").
		AddRow("writing", 5, "b")
	mock.ExpectQuery("SELECT .* FROM content_public_labels").
		WithArgs("writing", int32(5)).
		WillReturnRows(rows)
	mock.ExpectQuery("SELECT .* FROM content_label_status").
		WithArgs("writing", int32(5)).
		WillReturnRows(sqlmock.NewRows([]string{"item", "item_id", "label"}))
	mock.ExpectExec("INSERT IGNORE INTO content_public_labels").
		WithArgs("writing", int32(5), "c").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM content_public_labels").
		WithArgs("writing", int32(5), "a").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := cd.SetWritingPublicLabels(5, []string{"b", "c"}); err != nil {
		t.Fatalf("SetWritingPublicLabels: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
