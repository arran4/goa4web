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
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func TestForgotPassword_VerifiedEmail(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	mock.ExpectQuery("SystemGetLogin").WillReturnRows(sqlmock.NewRows([]string{"idusers", "passwd", "passwd_algorithm", "username"}).AddRow(1, "hash", "alg", "user"))
	mock.ExpectQuery("SystemListVerifiedEmailsByUserID").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}).AddRow(1, 1, "test@example.com", time.Now(), "", time.Now(), 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE ur.users_idusers = ? AND r.can_login = 1 LIMIT 1")).WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"column_1"}).AddRow(1))
	mock.ExpectQuery("SystemGetUserByEmail").WithArgs("test@example.com").WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "test@example.com", "user"))
	mock.ExpectQuery("GetPasswordResetByUser").WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("INSERT INTO pending_passwords").WillReturnResult(sqlmock.NewResult(1, 1))

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)

	form := url.Values{"username": {"user"}, "password": {"newpass"}}
	req := httptest.NewRequest("POST", "/forgot", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handlers.TaskHandler(forgotPasswordTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code = %v, want %v", rr.Code, http.StatusOK)
	}
	// Check content if possible (Template rendering might be mocked or we check if result is correct handler)
	// TaskHandler renders the template.
}

func TestForgotPassword_NoVerifiedEmail(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	mock.ExpectQuery("SystemGetLogin").WillReturnRows(sqlmock.NewRows([]string{"idusers", "passwd", "passwd_algorithm", "username"}).AddRow(1, "hash", "alg", "user"))
	mock.ExpectQuery("SystemListVerifiedEmailsByUserID").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"})) // No rows
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE ur.users_idusers = ? AND r.can_login = 1 LIMIT 1")).WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"column_1"}).AddRow(1))
	mock.ExpectQuery("GetPasswordResetByUser").WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("INSERT INTO pending_passwords").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("GetPasswordResetByCode").WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "passwd", "passwd_algorithm", "verification_code", "created_at", "verified_at"}).AddRow(100, 1, "hash", "alg", "code", time.Now(), nil))
	mock.ExpectExec("INSERT INTO admin_request_queue").WithArgs(int32(1), "users", "password_reset", int32(100), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)

	form := url.Values{"username": {"user"}, "password": {"newpass"}}
	req := httptest.NewRequest("POST", "/forgot", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handlers.TaskHandler(forgotPasswordTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code = %v, want %v", rr.Code, http.StatusOK)
	}
}

func TestResetPasswordAction(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	mock.ExpectQuery("GetPasswordResetByCode").WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "passwd", "passwd_algorithm", "verification_code", "created_at", "verified_at"}).AddRow(100, 1, nil, nil, "code", time.Now(), nil))
	mock.ExpectExec("INSERT INTO passwords").WithArgs(int32(1), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE pending_passwords SET verified_at").WithArgs(int32(100)).WillReturnResult(sqlmock.NewResult(1, 1))

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)

	form := url.Values{"code": {"code"}, "password": {"newpass"}}
	req := httptest.NewRequest("POST", "/reset", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handlers.TaskHandler(resetPasswordTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code = %v, want %v", rr.Code, http.StatusOK)
	}
	if body := rr.Body.String(); !strings.Contains(body, "url=/login") {
		t.Errorf("body does not contain refresh to /login: %s", body)
	}
}
