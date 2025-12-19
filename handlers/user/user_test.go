package user

import (
	"context"
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/arran4/goa4web/handlers"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	logProv "github.com/arran4/goa4web/internal/email/log"
	"time"
)

func newEmailReg() *email.Registry {
	r := email.NewRegistry()
	logProv.Register(r)
	return r
}

var (
	store       *sessions.CookieStore
	sessionName = "my-session"
)

// helper to setup session cookie
func newRequestWithSession(method, target string, values map[string]interface{}) (*http.Request, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(method, target, nil)
	w := httptest.NewRecorder()
	sess, _ := store.Get(r, sessionName)
	for k, v := range values {
		sess.Values[k] = v
	}
	sess.Save(r, w)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}
	return r, httptest.NewRecorder()
}

func TestUserEmailTestAction_NoProvider(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.EmailProvider = ""
	conn, mock, _ := sqlmock.New()
	defer conn.Close()
	queries := db.New(conn)
	mock.ExpectQuery("SELECT u.idusers, ue.email, u.username").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).AddRow(1, "e", "u", nil))
	mock.ExpectQuery("SELECT id, user_id, email").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}).AddRow(1, 1, "e", nil, nil, nil, 100))
	req := httptest.NewRequest("POST", "/email", nil)
	ctx := req.Context()
	reg := newEmailReg()
	p, _ := reg.ProviderFromConfig(cfg)
	cd := common.NewCoreData(ctx, queries, cfg, common.WithEmailProvider(p))
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(testMailTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	want := url.QueryEscape(ErrMailNotConfigured.Error())
	if req.URL.RawQuery != "error="+want {
		t.Fatalf("query=%q", req.URL.RawQuery)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "<a href=") {
		t.Fatalf("body=%q", body)
	}
}

func TestUserEmailTestAction_WithProvider(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.EmailProvider = "log"

	conn, mock, _ := sqlmock.New()
	defer conn.Close()
	queries := db.New(conn)
	mock.ExpectQuery("SELECT u.idusers, ue.email, u.username").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).AddRow(1, "e", "u", nil))
	mock.ExpectQuery("SELECT id, user_id, email").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}).AddRow(1, 1, "e", nil, nil, nil, 100))
	req := httptest.NewRequest("POST", "/email", nil)
	ctx := req.Context()
	reg := newEmailReg()
	p, _ := reg.ProviderFromConfig(cfg)
	cd := common.NewCoreData(ctx, queries, cfg, common.WithEmailProvider(p))
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(testMailTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestUserEmailPage_ShowError(t *testing.T) {
	conn, mock, _ := sqlmock.New()
	defer conn.Close()
	queries := db.New(conn)
	mock.ExpectQuery("SELECT u.idusers, ue.email, u.username").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).AddRow(1, "e", "u", nil))
	mock.ExpectQuery("SELECT id, user_id, email").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}).AddRow(1, 1, "e", nil, nil, nil, 100))
	req := httptest.NewRequest("GET", "/usr/email?error=missing", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	userEmailPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "missing") {
		t.Fatalf("body=%q", rr.Body.String())
	}
}

func TestUserEmailPage_NoUnverified(t *testing.T) {
	conn, mock, _ := sqlmock.New()
	defer conn.Close()
	queries := db.New(conn)
	mock.ExpectQuery("SELECT u.idusers, ue.email, u.username").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).AddRow(1, "e", "u", nil))
	mock.ExpectQuery("SELECT id, user_id, email").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}).AddRow(1, 1, "e", time.Now(), nil, nil, 100))

	req := httptest.NewRequest("GET", "/usr/email", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	userEmailPage(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, "Unverified Emails") {
		t.Fatalf("unverified section should be hidden: %q", body)
	}
	if !strings.Contains(body, "Verified Emails") {
		t.Fatalf("missing verified section: %q", body)
	}
}

func TestUserEmailPage_NoVerified(t *testing.T) {
	conn, mock, _ := sqlmock.New()
	defer conn.Close()
	queries := db.New(conn)
	mock.ExpectQuery("SELECT u.idusers, ue.email, u.username").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).AddRow(1, "e", "u", nil))
	mock.ExpectQuery("SELECT id, user_id, email").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}))

	req := httptest.NewRequest("GET", "/usr/email", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	userEmailPage(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "No verified emails") {
		t.Fatalf("missing warning message: %q", body)
	}
	if strings.Contains(body, "Unverified Emails") {
		t.Fatalf("unexpected unverified section: %q", body)
	}
}

func TestUserLangSaveAllActionPage_NewPref(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	store = sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = sessionName
	core.Store = store
	core.SessionName = sessionName

	form := url.Values{}
	form.Set("dothis", "Save all")
	form.Set("language1", "on")
	form.Set("defaultLanguage", "2")

	req := httptest.NewRequest("POST", "/usr/lang", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	sess, _ := store.Get(req, sessionName)
	sess.Values["UID"] = int32(1)
	w := httptest.NewRecorder()
	sess.Save(req, w)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}
	rr := httptest.NewRecorder()

	ctx := req.Context()
	cfg := config.NewRuntimeConfig()
	cfg.PageSizeDefault = 15
	cd := common.NewCoreData(ctx, queries, cfg, common.WithSession(sess))
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rows := sqlmock.NewRows([]string{"id", "nameof"}).AddRow(1, "en").AddRow(2, "fr")
	mock.ExpectExec("DELETE FROM user_language").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, nameof\nFROM language")).WillReturnRows(rows)
	mock.ExpectExec("INSERT INTO user_language").WithArgs(int32(1), int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT idpreferences").WithArgs(int32(1)).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("INSERT INTO preferences").WithArgs(int32(2), int32(1), int32(cfg.PageSizeDefault), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

	saveAllTask.Action(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}

func TestUserLangSaveLanguagesActionPage(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	store = sessions.NewCookieStore([]byte("test"))

	form := url.Values{}
	form.Set("dothis", "Save languages")
	form.Set("language1", "on")

	req := httptest.NewRequest("POST", "/usr/lang", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	sess, _ := store.Get(req, sessionName)
	sess.Values["UID"] = int32(1)
	w := httptest.NewRecorder()
	sess.Save(req, w)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}
	rr := httptest.NewRecorder()

	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithSession(sess))
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rows := sqlmock.NewRows([]string{"id", "nameof"}).AddRow(1, "en")
	mock.ExpectExec("DELETE FROM user_language").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, nameof\nFROM language")).WillReturnRows(rows)
	mock.ExpectExec("INSERT INTO user_language").WithArgs(int32(1), int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))

	saveLanguagesTask.Action(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}

func TestUserLangSaveLanguageActionPage_UpdatePref(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	store = sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = sessionName

	form := url.Values{}
	form.Set("dothis", "Save language")
	form.Set("defaultLanguage", "2")

	req := httptest.NewRequest("POST", "/usr/lang", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	sess, _ := store.Get(req, sessionName)
	sess.Values["UID"] = int32(1)
	w := httptest.NewRecorder()
	sess.Save(req, w)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}
	rr := httptest.NewRecorder()
	cfg := config.NewRuntimeConfig()
	cfg.PageSizeDefault = 15

	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, cfg, common.WithSession(sess))
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	prefRows := sqlmock.NewRows([]string{"idpreferences", "language_id", "users_idusers", "emailforumupdates", "page_size", "auto_subscribe_replies", "timezone"}).
		AddRow(1, 1, 1, nil, cfg.PageSizeDefault, true, nil)
	mock.ExpectQuery("SELECT idpreferences").WithArgs(int32(1)).WillReturnRows(prefRows)
	mock.ExpectExec("UPDATE preferences").WithArgs(int32(2), int32(cfg.PageSizeDefault), sqlmock.AnyArg(), int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))

	saveLanguageTask.Action(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}
