package auth

import (
	"context"
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

func TestForgotPasswordRateLimit(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	mock.ExpectQuery("SystemGetLogin").WillReturnRows(sqlmock.NewRows([]string{"idusers", "passwd", "passwd_algorithm", "username"}).AddRow(1, "", "", "u"))
	mock.ExpectQuery("SystemListVerifiedEmailsByUserID").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}).AddRow(1, 1, "a@test.com", time.Now(), nil, nil, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE ur.users_idusers = ? AND r.can_login = 1 LIMIT 1")).WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"column_1"}).AddRow(1))
	mock.ExpectQuery("SystemGetUserByEmail").WithArgs("a@test.com").WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "a@test.com", "u"))
	resetRows := sqlmock.NewRows([]string{"id", "user_id", "passwd", "passwd_algorithm", "verification_code", "created_at", "verified_at"}).
		AddRow(1, 1, "hash", "alg", "code", time.Now(), nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, passwd")).WithArgs(int32(1), sqlmock.AnyArg()).WillReturnRows(resetRows)

	form := url.Values{"username": {"u"}, "password": {"pw"}}
	req := httptest.NewRequest(http.MethodPost, "/forgot", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handlers.TaskHandler(forgotPasswordTask)(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestForgotPasswordReplaceOld(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	mock.ExpectQuery("SystemGetLogin").WillReturnRows(sqlmock.NewRows([]string{"idusers", "passwd", "passwd_algorithm", "username"}).AddRow(1, "", "", "u"))
	mock.ExpectQuery("SystemListVerifiedEmailsByUserID").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}).AddRow(1, 1, "a@test.com", time.Now(), nil, nil, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE ur.users_idusers = ? AND r.can_login = 1 LIMIT 1")).WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"column_1"}).AddRow(1))
	mock.ExpectQuery("SystemGetUserByEmail").WithArgs("a@test.com").WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "a@test.com", "u"))
	oldRows := sqlmock.NewRows([]string{"id", "user_id", "passwd", "passwd_algorithm", "verification_code", "created_at", "verified_at"}).
		AddRow(1, 1, "hash", "alg", "code", time.Now().Add(-25*time.Hour), nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, passwd")).WithArgs(int32(1), sqlmock.AnyArg()).WillReturnRows(oldRows)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM pending_passwords WHERE id = ?")).WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO pending_passwords").WillReturnResult(sqlmock.NewResult(2, 1))

	form := url.Values{"username": {"u"}, "password": {"pw"}}
	req := httptest.NewRequest(http.MethodPost, "/forgot", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)
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
