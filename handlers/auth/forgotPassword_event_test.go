package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
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
	mock.ExpectExec("INSERT INTO pending_passwords").WillReturnResult(sqlmock.NewResult(1, 1))

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	cd := common.NewCoreData(context.Background(), q, common.WithEvent(evt))
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)
	ctx = ctx

	form := url.Values{"username": {"u"}, "password": {"pw"}}
	req := httptest.NewRequest("POST", "/forgot", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	forgotPasswordTask.Action(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if _, ok := evt.Data["reset"]; !ok {
		t.Fatalf("missing reset data")
	}
	if _, ok := evt.Data["ResetURL"]; !ok {
		t.Fatalf("missing ResetURL data")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
