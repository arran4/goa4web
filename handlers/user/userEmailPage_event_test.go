package user

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
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
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
	mock.ExpectQuery("SELECT u.idusers").WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).
			AddRow(1, nil, "alice"))

	store = sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test"
	sess, _ := store.Get(httptest.NewRequest(http.MethodGet, "http://example.com", nil), core.SessionName)
	sess.Values["UID"] = int32(1)
	w := httptest.NewRecorder()
	_ = sess.Save(httptest.NewRequest(http.MethodGet, "http://example.com", nil), w)

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	ctx := context.Background()
	cd := common.NewCoreData(ctx, q, common.WithSession(sess), common.WithEvent(evt), config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" }))
	cd.UserID = 1
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	req := httptest.NewRequest("POST", "http://example.com/usr/email", strings.NewReader("new_email=a@example.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handlers.TaskHandler(addEmailTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if _, ok := evt.Data["URL"]; !ok {
		t.Fatalf("missing URL event data: %+v", evt.Data)
	}
	if evt.Data["Username"] != "alice" {
		t.Fatalf("username not set: %+v", evt.Data)
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

	store = sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test"
	sess, _ := store.Get(httptest.NewRequest(http.MethodGet, "http://example.com", nil), core.SessionName)
	sess.Values["UID"] = int32(1)
	w := httptest.NewRecorder()
	_ = sess.Save(httptest.NewRequest(http.MethodGet, "http://example.com", nil), w)

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	ctx := context.Background()
	cd := common.NewCoreData(ctx, q, common.WithSession(sess), common.WithEvent(evt), config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" }))
	cd.UserID = 1

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

	sess.Values = map[interface{}]interface{}{"UID": int32(1)}
	core.Store = store
	core.SessionName = "test"

	form := url.Values{"code": {"code"}}
	req := httptest.NewRequest(http.MethodPost, "/usr/email/verify", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
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

func TestResendVerificationEmailTaskEventData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	mock.ExpectQuery("SELECT id, user_id, email").WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}).AddRow(1, 1, "a@example.com", nil, nil, nil, 0))
	mock.ExpectExec("UPDATE user_emails SET").WillReturnResult(sqlmock.NewResult(1, 1))

	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test"

	req := httptest.NewRequest("POST", "http://example.com/usr/email/resend", strings.NewReader("id=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	sess, _ := store.Get(req, core.SessionName)
	sess.Values["UID"] = int32(1)
	w := httptest.NewRecorder()
	_ = sess.Save(req, w)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	ctx := context.Background()
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	cd := common.NewCoreData(ctx, q, common.WithSession(sess), common.WithEvent(evt), config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" }))
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handlers.TaskHandler(resendVerificationEmailTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if _, ok := evt.Data["page"]; !ok {
		t.Fatalf("missing page event data: %+v", evt.Data)
	}
	if _, ok := evt.Data["email"]; !ok {
		t.Fatalf("missing email event data: %+v", evt.Data)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
