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

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func setupSubscriptionTaskTest(t *testing.T, userID int, body url.Values) (*httptest.ResponseRecorder, *http.Request, *sql.DB, sqlmock.Sqlmock) {
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
	req := httptest.NewRequest("POST", "/admin/user/"+strconv.Itoa(userID)+"/subscriptions", reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req = mux.SetURLVars(req, map[string]string{"user": strconv.Itoa(userID)})
	cfg := config.NewRuntimeConfig()
	q := db.New(conn)
	cd := common.NewCoreData(req.Context(), q, cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	return httptest.NewRecorder(), req, conn, mock
}

func TestAddUserSubscriptionTask_UsesURLParam(t *testing.T) {
	body := url.Values{"pattern": {"/foo"}, "method": {"internal"}}
	rr, req, conn, mock := setupSubscriptionTaskTest(t, 9, body)
	defer conn.Close()
	mock.ExpectExec("INSERT").WithArgs(int32(9), "/foo", "internal").WillReturnResult(sqlmock.NewResult(0, 1))
	if err, ok := addUserSubscriptionTask.Action(rr, req).(error); ok && err != nil {
		t.Fatalf("Action: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestUpdateUserSubscriptionTask_UsesURLParam(t *testing.T) {
	body := url.Values{"id": {"3"}, "pattern": {"/bar"}, "method": {"email"}}
	rr, req, conn, mock := setupSubscriptionTaskTest(t, 4, body)
	defer conn.Close()
	mock.ExpectExec("UPDATE").WithArgs("/bar", "email", int32(4), int32(3)).WillReturnResult(sqlmock.NewResult(0, 1))
	if err, ok := updateUserSubscriptionTask.Action(rr, req).(error); ok && err != nil {
		t.Fatalf("Action: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestDeleteUserSubscriptionTask_UsesURLParam(t *testing.T) {
	body := url.Values{"id": {"5"}}
	rr, req, conn, mock := setupSubscriptionTaskTest(t, 11, body)
	defer conn.Close()
	mock.ExpectExec("DELETE").WithArgs(int32(11), int32(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	if err, ok := deleteUserSubscriptionTask.Action(rr, req).(error); ok && err != nil {
		t.Fatalf("Action: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}
