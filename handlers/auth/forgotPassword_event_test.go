package auth

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

func TestForgotPasswordEventData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	mock.ExpectQuery("GetUserByUsername").WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "a@test.com", "u"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, passwd")).WithArgs(int32(1), sqlmock.AnyArg()).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("INSERT INTO pending_passwords").WillReturnResult(sqlmock.NewResult(1, 1))

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig(), common.WithEvent(evt))
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)

	form := url.Values{"username": {"u"}, "password": {"pw"}}
	req := httptest.NewRequest("POST", "/forgot", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handlers.TaskHandler(forgotPasswordTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if _, ok := evt.Data["Username"]; !ok {
		t.Fatalf("missing Username data")
	}
	if _, ok := evt.Data["Code"]; !ok {
		t.Fatalf("missing Code data")
	}
	if _, ok := evt.Data["Username"]; !ok {
		t.Fatalf("missing Username data")
	}
	if _, ok := evt.Data["Code"]; !ok {
		t.Fatalf("missing Code data")
	}
	if _, ok := evt.Data["ResetURL"]; !ok {
		t.Fatalf("missing ResetURL data")
	}
	if _, ok := evt.Data["UserURL"]; !ok {
		t.Fatalf("missing UserURL data")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
