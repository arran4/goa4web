package user

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/gorilla/sessions"
)

func TestAddEmailTaskEventData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	mock.ExpectQuery("SELECT id, user_id, email").WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("INSERT INTO user_emails").WillReturnResult(sqlmock.NewResult(1, 1))

	store := sessions.NewCookieStore([]byte("test"))
	sess := sessions.NewSession(store, "test")
	sess.Values = map[interface{}]interface{}{"UID": int32(1)}
	core.Store = store
	core.SessionName = "test"

	evt := &eventbus.Event{Data: map[string]any{}}
	ctx := context.WithValue(context.Background(), consts.KeyQueries, q)
	cd := common.NewCoreData(ctx, q, common.WithSession(sess), common.WithEvent(evt))
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	req := httptest.NewRequest("POST", "http://example.com/usr/email", strings.NewReader("new_email=a@example.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	addEmailTask.Action(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if _, ok := evt.Data["URL"]; !ok {
		t.Fatalf("missing URL event data: %+v", evt.Data)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestVerifyRemovesDuplicates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)

	rows := sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}).
		AddRow(1, 1, "a@example.com", nil, "code", time.Now().Add(time.Hour), 0)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority\nFROM user_emails\nWHERE last_verification_code = ?")).
		WithArgs(sql.NullString{String: "code", Valid: true}).
		WillReturnRows(rows)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE user_emails\nSET verified_at = ?, last_verification_code = NULL, verification_expires_at = NULL\nWHERE id = ?")).
		WithArgs(sqlmock.AnyArg(), int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM user_emails WHERE email = ? AND id != ?")).
		WithArgs("a@example.com", int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest(http.MethodGet, "/usr/email/verify?code=code", nil)
	ctx := context.WithValue(req.Context(), consts.KeyQueries, q)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	userEmailVerifyCodePage(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
