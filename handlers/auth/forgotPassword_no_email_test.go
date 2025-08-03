package auth

import (
	"context"
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
	"github.com/arran4/goa4web/internal/db"
)

func TestForgotPasswordNoEmail(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	mock.ExpectQuery("SystemGetUserByUsername").WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).AddRow(1, "", "u", nil))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE ur.users_idusers = ? AND r.can_login = 1 LIMIT 1")).WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"column_1"}).AddRow(1))

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
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
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestEmailAssociationRequestTask(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	mock.ExpectQuery("SystemGetUserByUsername").WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).AddRow(1, "", "u", nil))
	mock.ExpectExec("INSERT INTO admin_request_queue").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO admin_request_comments").WillReturnResult(sqlmock.NewResult(1, 1))

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)

	form := url.Values{"username": {"u"}, "email": {"a@test.com"}, "reason": {"help"}}
	req := httptest.NewRequest("POST", "/forgot", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handlers.TaskHandler(emailAssociationRequestTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
