package common_test

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestThreadReadMarker(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 1

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO content_read_markers")).
		WithArgs("thread", int32(2), cd.UserID, int32(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := cd.SetThreadReadMarker(2, 5); err != nil {
		t.Fatalf("SetThreadReadMarker: %v", err)
	}

	rows := sqlmock.NewRows([]string{"item", "item_id", "user_id", "last_comment_id"}).
		AddRow("thread", 2, cd.UserID, 5)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT item, item_id, user_id, last_comment_id FROM content_read_markers")).
		WithArgs("thread", int32(2), cd.UserID).
		WillReturnRows(rows)

	cid, err := cd.ThreadReadMarker(2)
	if err != nil {
		t.Fatalf("ThreadReadMarker: %v", err)
	}
	if cid != 5 {
		t.Fatalf("last comment %d, want 5", cid)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
