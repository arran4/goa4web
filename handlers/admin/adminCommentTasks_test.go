package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func setupCommentTest(t *testing.T, commentID int, body url.Values) (*httptest.ResponseRecorder, *http.Request, *sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	mock.MatchExpectationsInOrder(false)
	var reader *strings.Reader
	if body != nil {
		reader = strings.NewReader(body.Encode())
	} else {
		reader = strings.NewReader("")
	}
	req := httptest.NewRequest("POST", "/admin/comment/"+strconv.Itoa(commentID), reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req = mux.SetURLVars(req, map[string]string{"comment": strconv.Itoa(commentID)})
	cfg := config.NewRuntimeConfig()
	q := db.New(conn)
	cd := common.NewCoreData(req.Context(), q, cfg, common.WithSelectionsFromRequest(req))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	return httptest.NewRecorder(), req, conn, mock
}

func TestDeleteCommentTask_UsesURLParam(t *testing.T) {
	rr, req, conn, mock := setupCommentTest(t, 15, nil)
	defer conn.Close()
	rows := sqlmock.NewRows([]string{"idcomments", "forumthread_id", "users_idusers", "language_idlanguage", "written", "text", "deleted_at", "last_index", "username", "is_owner"}).
		AddRow(15, 2, 3, 1, time.Now(), "body", nil, nil, "user", true)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
	mock.ExpectExec("UPDATE").WithArgs(sqlmock.AnyArg(), int32(15)).WillReturnResult(sqlmock.NewResult(0, 1))
	if err, ok := deleteCommentTask.Action(rr, req).(error); ok && err != nil {
		t.Fatalf("Action: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestEditCommentTask_UsesURLParam(t *testing.T) {
	body := url.Values{"replytext": {"updated"}}
	rr, req, conn, mock := setupCommentTest(t, 22, body)
	defer conn.Close()
	rows := sqlmock.NewRows([]string{"idcomments", "forumthread_id", "users_idusers", "language_idlanguage", "written", "text", "deleted_at", "last_index", "username", "is_owner"}).
		AddRow(22, 2, 3, 1, time.Now(), "body", nil, nil, "user", true)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
	mock.ExpectExec("UPDATE").WithArgs(sqlmock.AnyArg(), int32(22)).WillReturnResult(sqlmock.NewResult(0, 1))
	if err, ok := editCommentTask.Action(rr, req).(error); ok && err != nil {
		t.Fatalf("Action: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestBanCommentTask_UsesURLParam(t *testing.T) {
	rr, req, conn, mock := setupCommentTest(t, 33, nil)
	defer conn.Close()
	rows := sqlmock.NewRows([]string{"idcomments", "forumthread_id", "users_idusers", "language_idlanguage", "written", "text", "deleted_at", "last_index", "username", "is_owner"}).
		AddRow(33, 2, 3, 1, time.Now(), "body", nil, nil, "user", true)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
	mock.ExpectExec("INSERT").WithArgs(int32(33), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE").WithArgs(sqlmock.AnyArg(), int32(33)).WillReturnResult(sqlmock.NewResult(0, 1))
	if err, ok := banCommentTask.Action(rr, req).(error); ok && err != nil {
		t.Fatalf("Action: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}
