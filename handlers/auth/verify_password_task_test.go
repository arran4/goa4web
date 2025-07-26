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
	dbpkg "github.com/arran4/goa4web/internal/db"
)

func TestVerifyPasswordUsesHashedCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)

	hashed := HashResetCode("abc")
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "passwd", "passwd_algorithm", "verification_code", "created_at", "verified_at"}).
		AddRow(1, 2, "hash", "alg", hashed, now, nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, passwd")).
		WithArgs(hashed, sqlmock.AnyArg()).
		WillReturnRows(rows)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE pending_passwords SET verified_at = NOW() WHERE id = ?")).
		WithArgs(int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO passwords")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	cd := common.NewCoreData(context.Background(), q, common.WithConfig(config.AppRuntimeConfig))
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)

	form := url.Values{"id": {"1"}, "code": {"abc"}}
	req := httptest.NewRequest(http.MethodPost, "/verify", strings.NewReader(form.Encode()))
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
