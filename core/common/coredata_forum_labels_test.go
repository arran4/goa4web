package common_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/arran4/goa4web/config"
	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestSetTopicPublicLabels(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 2

	rows := sqlmock.NewRows([]string{"forumtopic_idforumtopic", "label"}).
		AddRow(1, "foo").
		AddRow(1, "bar")
	mock.ExpectQuery("SELECT .* FROM forumtopic_public_labels").
		WithArgs(int32(1)).
		WillReturnRows(rows)
	mock.ExpectQuery("SELECT .* FROM content_label_status").
		WithArgs("forumtopic", int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"item", "item_id", "label"}))
	mock.ExpectExec("INSERT IGNORE INTO forumtopic_public_labels").
		WithArgs(int32(1), "baz").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM forumtopic_public_labels").
		WithArgs(int32(1), "foo").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := cd.SetTopicPublicLabels(1, []string{"bar", "baz"}); err != nil {
		t.Fatalf("SetTopicPublicLabels: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSetTopicPrivateLabels(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 2

	rows := sqlmock.NewRows([]string{"forumtopic_idforumtopic", "users_idusers", "label", "invert"}).
		AddRow(1, 2, "one", false).
		AddRow(1, 2, "two", false)
	mock.ExpectQuery("SELECT .* FROM forumtopic_private_labels").
		WithArgs(int32(1), int32(2)).
		WillReturnRows(rows)
	mock.ExpectExec("INSERT IGNORE INTO forumtopic_private_labels").
		WithArgs(int32(1), int32(2), "three", false).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM forumtopic_private_labels").
		WithArgs(int32(1), int32(2), "one").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := cd.SetTopicPrivateLabels(1, []string{"two", "three"}); err != nil {
		t.Fatalf("SetTopicPrivateLabels: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
