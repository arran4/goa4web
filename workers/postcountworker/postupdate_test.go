package postcountworker

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/internal/db"
)

func TestPostUpdate(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)
	mock.ExpectExec("AdminRecalculateForumThreadByIdMetaData").
		WithArgs(int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SystemRebuildForumTopicMetaByID").
		WithArgs(int32(2)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := PostUpdate(context.Background(), q, 1, 2); err != nil {
		t.Fatalf("PostUpdate: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}
