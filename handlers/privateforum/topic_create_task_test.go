package privateforum

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func TestPrivateTopicCreateTask_GrantsBeforeComment(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)

	mock.ExpectQuery("SELECT 1 FROM grants").
		WithArgs(int64(1), "privateforum", "topic", "create", nil, int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectExec("INSERT INTO forumtopic").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO forumthread").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec("INSERT INTO grants").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO grants").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO grants").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO grants").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO comments").WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectExec("INSERT INTO subscriptions").WillReturnResult(sqlmock.NewResult(0, 1))

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 1

	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("body=hello"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	res := privateTopicCreateTask.Action(w, req)
	if err, ok := res.(error); ok {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := res.(handlers.RefreshDirectHandler); !ok {
		t.Fatalf("unexpected result: %#v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
