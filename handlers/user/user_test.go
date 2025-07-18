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

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers/common"
	dbpkg "github.com/arran4/goa4web/internal/db"
	logProv "github.com/arran4/goa4web/internal/email/log"
)

func init() { logProv.Register() }

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
	config.AppRuntimeConfig.EmailProvider = ""
	db, mock, _ := sqlmock.New()
	defer db.Close()
	queries := dbpkg.New(db)
	mock.ExpectQuery("SELECT u.idusers, ue.email, u.username").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "e", "u"))
	mock.ExpectQuery("SELECT id, user_id, email").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}).AddRow(1, 1, "e", nil, nil, nil, 100))
	req := httptest.NewRequest("POST", "/email", nil)
	ctx := context.WithValue(req.Context(), common.KeyQueries, queries)
	cd := corecommon.NewCoreData(ctx, queries)
	cd.UserID = 1
	ctx = context.WithValue(ctx, common.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	userEmailTestActionPage(rr, req)

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
	config.AppRuntimeConfig.EmailProvider = "log"

	db, mock, _ := sqlmock.New()
	defer db.Close()
	queries := dbpkg.New(db)
	mock.ExpectQuery("SELECT u.idusers, ue.email, u.username").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "e", "u"))
	mock.ExpectQuery("SELECT id, user_id, email").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}).AddRow(1, 1, "e", nil, nil, nil, 100))
	req := httptest.NewRequest("POST", "/email", nil)
	ctx := context.WithValue(req.Context(), common.KeyQueries, queries)
	cd := corecommon.NewCoreData(ctx, queries)
	cd.UserID = 1
	ctx = context.WithValue(ctx, common.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	userEmailTestActionPage(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/usr/email" {
		t.Fatalf("location=%q", loc)
	}
}

func TestUserEmailPage_ShowError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	queries := dbpkg.New(db)
	mock.ExpectQuery("SELECT u.idusers, ue.email, u.username").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "e", "u"))
	mock.ExpectQuery("SELECT id, user_id, email").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "verified_at", "last_verification_code", "verification_expires_at", "notification_priority"}).AddRow(1, 1, "e", nil, nil, nil, 100))
	req := httptest.NewRequest("GET", "/usr/email?error=missing", nil)
	ctx := context.WithValue(req.Context(), common.KeyQueries, queries)
	cd := corecommon.NewCoreData(ctx, queries)
	cd.UserID = 1
	ctx = context.WithValue(ctx, common.KeyCoreData, cd)
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

func TestUserLangSaveAllActionPage_NewPref(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)
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

	ctx := context.WithValue(req.Context(), common.KeyQueries, queries)
	cd := corecommon.NewCoreData(ctx, queries, corecommon.WithSession(sess))
	cd.UserID = 1
	ctx = context.WithValue(ctx, common.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rows := sqlmock.NewRows([]string{"idlanguage", "nameof"}).AddRow(1, "en").AddRow(2, "fr")
	mock.ExpectExec("DELETE FROM user_language").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idlanguage, nameof\nFROM language")).WillReturnRows(rows)
	mock.ExpectExec("INSERT INTO user_language").WithArgs(int32(1), int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT idpreferences").WithArgs(int32(1)).WillReturnError(sql.ErrNoRows)
	config.AppRuntimeConfig.PageSizeDefault = 15
	mock.ExpectExec("INSERT INTO preferences").WithArgs(int32(2), int32(1), int32(config.AppRuntimeConfig.PageSizeDefault)).WillReturnResult(sqlmock.NewResult(1, 1))

	userLangSaveAllActionPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}

func TestUserLangSaveLanguagesActionPage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)
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

	ctx := context.WithValue(req.Context(), common.KeyQueries, queries)
	cd := corecommon.NewCoreData(ctx, queries, corecommon.WithSession(sess))
	cd.UserID = 1
	ctx = context.WithValue(ctx, common.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rows := sqlmock.NewRows([]string{"idlanguage", "nameof"}).AddRow(1, "en")
	mock.ExpectExec("DELETE FROM user_language").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idlanguage, nameof\nFROM language")).WillReturnRows(rows)
	mock.ExpectExec("INSERT INTO user_language").WithArgs(int32(1), int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))

	userLangSaveLanguagesActionPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}

func TestUserLangSaveLanguageActionPage_UpdatePref(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)
	config.AppRuntimeConfig.PageSizeDefault = 15
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

	ctx := context.WithValue(req.Context(), common.KeyQueries, queries)
	cd := corecommon.NewCoreData(ctx, queries, corecommon.WithSession(sess))
	cd.UserID = 1
	ctx = context.WithValue(ctx, common.KeyCoreData, cd)
	req = req.WithContext(ctx)

	prefRows := sqlmock.NewRows([]string{"idpreferences", "language_idlanguage", "users_idusers", "emailforumupdates", "page_size", "auto_subscribe_replies"}).
		AddRow(1, 1, 1, nil, config.AppRuntimeConfig.PageSizeDefault, true)
	mock.ExpectQuery("SELECT idpreferences").WithArgs(int32(1)).WillReturnRows(prefRows)
	mock.ExpectExec("UPDATE preferences").WithArgs(int32(2), int32(config.AppRuntimeConfig.PageSizeDefault), int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))

	userLangSaveLanguagePreferenceActionPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}
