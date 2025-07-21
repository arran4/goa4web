package user

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

func TestUserEmailVerifyCodePage_Forbidden(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	code := "abc"
	rows := sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}).
		AddRow(1, 1, "e@example.com", nil, code, nil, 0)
	mock.ExpectQuery("SELECT id, user_id, email").WithArgs(sql.NullString{String: code, Valid: true}).WillReturnRows(rows)

	store := sessions.NewCookieStore([]byte("test"))
	sess := sessions.NewSession(store, "test")
	sess.Values = map[interface{}]interface{}{"UID": int32(2)}
	core.Store = store
	core.SessionName = "test"

	ctx := context.Background()
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	cd := common.NewCoreData(ctx, q, common.WithSession(sess))
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	req := httptest.NewRequest("GET", "/usr/email/verify?code="+code, nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	userEmailVerifyCodePage(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUserEmailVerifyCodePage_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	code := "xyz"
	rows := sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}).
		AddRow(1, 1, "e@example.com", nil, code, nil, 0)
	mock.ExpectQuery("SELECT id, user_id, email").WithArgs(sql.NullString{String: code, Valid: true}).WillReturnRows(rows)
	mock.ExpectExec("UPDATE user_emails").WithArgs(sqlmock.AnyArg(), int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))

	store := sessions.NewCookieStore([]byte("test"))
	sess := sessions.NewSession(store, "test")
	sess.Values = map[interface{}]interface{}{"UID": int32(1)}
	core.Store = store
	core.SessionName = "test"

	ctx := context.Background()
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	cd := common.NewCoreData(ctx, q, common.WithSession(sess))
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	req := httptest.NewRequest("GET", "/usr/email/verify?code="+code, nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	userEmailVerifyCodePage(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/usr/email" {
		t.Fatalf("location=%q", loc)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
