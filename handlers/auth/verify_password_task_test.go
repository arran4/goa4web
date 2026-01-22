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

func TestVerifyPasswordAction_Success(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	pwHash, alg, _ := common.HashPassword("pw")
	rows := sqlmock.NewRows([]string{"id", "user_id", "passwd", "passwd_algorithm", "verification_code", "created_at", "verified_at"}).
		AddRow(1, 1, sql.NullString{String: pwHash, Valid: true}, sql.NullString{String: alg, Valid: true}, "code", time.Now(), nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, passwd")).WithArgs("code", sqlmock.AnyArg()).WillReturnRows(rows)
	// CoreData.VerifyPasswordReset does a role check
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE ur.users_idusers = ? AND r.can_login = 1 LIMIT 1")).WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"column_1"}).AddRow(1))

	mock.ExpectExec(regexp.QuoteMeta("UPDATE pending_passwords SET verified_at = NOW() WHERE id = ?")).WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO passwords (users_idusers, passwd, passwd_algorithm) VALUES (?, ?, ?)")).
		WithArgs(int32(1), sql.NullString{String: pwHash, Valid: true}, sql.NullString{String: alg, Valid: true}).
		WillReturnResult(sqlmock.NewResult(1, 1))

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)
	form := url.Values{"id": {"1"}, "code": {"code"}, "password": {"pw"}}
	req := httptest.NewRequest(http.MethodPost, "/login/verify", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(verifyPasswordTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestVerifyPasswordAction_InvalidPassword(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	pwHash, alg, _ := common.HashPassword("pw")
	rows := sqlmock.NewRows([]string{"id", "user_id", "passwd", "passwd_algorithm", "verification_code", "created_at", "verified_at"}).
		AddRow(1, 1, pwHash, alg, "code", time.Now(), nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, passwd")).WithArgs("code", sqlmock.AnyArg()).WillReturnRows(rows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE ur.users_idusers = ? AND r.can_login = 1 LIMIT 1")).WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"column_1"}).AddRow(1))

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)
	form := url.Values{"id": {"1"}, "code": {"code"}, "password": {"wrong"}}
	req := httptest.NewRequest(http.MethodPost, "/login/verify", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(verifyPasswordTask)(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
