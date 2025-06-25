package user

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/handlers/common"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/runtimeconfig"
)

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

func TestUserAdderMiddleware_ExpiredSession(t *testing.T) {
	store = sessions.NewCookieStore([]byte("test-key"))
	core.Store = store
	core.SessionName = sessionName
	core.Store = store
	core.SessionName = sessionName
	req, rr := newRequestWithSession("GET", "/", map[string]interface{}{
		"UID":        int32(1),
		"ExpiryTime": time.Now().Add(-time.Hour).Unix(),
	})
	ctx := context.WithValue(req.Context(), common.KeyQueries, dbpkg.New(nil))
	req = req.WithContext(ctx)

	called := false
	handler := UserAdderMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	handler.ServeHTTP(rr, req)
	if rr.Result().StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected redirect, got %d", rr.Result().StatusCode)
	}
	if loc := rr.Result().Header.Get("Location"); !strings.HasPrefix(loc, "/login") {
		t.Errorf("expected redirect to /login, got %s", loc)
	}
	if called {
		t.Errorf("handler should not be called")
	}
}

func TestUserAdderMiddleware_AttachesPrefs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	store = sessions.NewCookieStore([]byte("test-key"))

	req, rr := newRequestWithSession("GET", "/", map[string]interface{}{
		"UID":        int32(1),
		"ExpiryTime": time.Now().Add(time.Hour).Unix(),
	})

	queries := dbpkg.New(db)
	ctx := context.WithValue(req.Context(), common.KeyQueries, queries)
	req = req.WithContext(ctx)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT idusers, email, passwd, passwd_algorithm, username\nFROM users\nWHERE idusers = ?")).
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "passwd", "passwd_algorithm", "username"}).
			AddRow(1, "e", "p", "", "u"))
	mock.ExpectQuery("SELECT idpermissions, users_idusers, section, level FROM permissions WHERE users_idusers = ?").
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"idpermissions", "users_idusers", "section", "level"}).AddRow(1, 1, "all", "admin"))
	mock.ExpectQuery("SELECT idpreferences, language_idlanguage, users_idusers, emailforumupdates, page_size FROM preferences WHERE users_idusers = ?").
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"idpreferences", "language_idlanguage", "users_idusers", "emailforumupdates", "page_size"}).AddRow(1, 2, 1, false, 15))
	mock.ExpectQuery("SELECT iduserlang, users_idusers, language_idlanguage FROM userlang WHERE users_idusers = ?").
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"iduserlang", "users_idusers", "language_idlanguage"}).AddRow(1, 1, 2))

	var gotPerms []*dbpkg.Permission
	var gotPref *dbpkg.Preference
	var gotLangs []*dbpkg.Userlang

	handler := UserAdderMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPerms, _ = r.Context().Value(common.KeyPermissions).([]*dbpkg.Permission)
		gotPref, _ = r.Context().Value(common.KeyPreference).(*dbpkg.Preference)
		gotLangs, _ = r.Context().Value(common.KeyLanguages).([]*dbpkg.Userlang)
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d", rr.Code)
	}
	if len(gotPerms) != 1 || gotPerms[0].Level.String != "admin" {
		t.Errorf("permissions not attached")
	}
	if gotPref == nil || gotPref.LanguageIdlanguage != 2 {
		t.Errorf("preference missing")
	}
	if len(gotLangs) != 1 || gotLangs[0].LanguageIdlanguage != 2 {
		t.Errorf("languages missing")
	}
}

func TestUserEmailTestAction_NoProvider(t *testing.T) {
	os.Unsetenv("EMAIL_PROVIDER")
	runtimeconfig.AppRuntimeConfig.EmailProvider = ""
	req := httptest.NewRequest("POST", "/email", nil)
	ctx := context.WithValue(req.Context(), common.KeyUser, &dbpkg.User{Email: sql.NullString{String: "u@example.com", Valid: true}})
	ctx = context.WithValue(ctx, common.KeyCoreData, &common.CoreData{})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	userEmailTestActionPage(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status=%d", rr.Code)
	}
	want := url.QueryEscape(ErrMailNotConfigured)
	if loc := rr.Header().Get("Location"); !strings.Contains(loc, want) {
		t.Fatalf("location=%q", loc)
	}
}

func TestUserEmailTestAction_WithProvider(t *testing.T) {
	os.Setenv("EMAIL_PROVIDER", "log")
	runtimeconfig.AppRuntimeConfig.EmailProvider = "log"
	defer os.Unsetenv("EMAIL_PROVIDER")

	req := httptest.NewRequest("POST", "/email", nil)
	ctx := context.WithValue(req.Context(), common.KeyUser, &dbpkg.User{Email: sql.NullString{String: "u@example.com", Valid: true}})
	ctx = context.WithValue(ctx, common.KeyCoreData, &common.CoreData{})
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
	req := httptest.NewRequest("GET", "/usr/email?error=missing", nil)
	ctx := context.WithValue(req.Context(), common.KeyUser, &dbpkg.User{Email: sql.NullString{String: "u@example.com", Valid: true}})
	ctx = context.WithValue(ctx, common.KeyCoreData, &common.CoreData{})
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
	ctx = context.WithValue(ctx, common.KeySession, sess)
	ctx = context.WithValue(ctx, common.KeyCoreData, &common.CoreData{})
	req = req.WithContext(ctx)
	rows := sqlmock.NewRows([]string{"idlanguage", "nameof"}).AddRow(1, "en").AddRow(2, "fr")
	mock.ExpectExec("DELETE FROM userlang").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idlanguage, nameof\nFROM language")).WillReturnRows(rows)
	mock.ExpectExec("INSERT INTO userlang").WithArgs(int32(1), int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT idpreferences").WithArgs(int32(1)).WillReturnError(sql.ErrNoRows)
	runtimeconfig.AppRuntimeConfig.PageSizeDefault = 15
	mock.ExpectExec("INSERT INTO preferences").WithArgs(int32(2), int32(1), int32(runtimeconfig.AppRuntimeConfig.PageSizeDefault)).WillReturnResult(sqlmock.NewResult(1, 1))

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
	ctx = context.WithValue(ctx, common.KeySession, sess)
	ctx = context.WithValue(ctx, common.KeyCoreData, &common.CoreData{})
	req = req.WithContext(ctx)

	rows := sqlmock.NewRows([]string{"idlanguage", "nameof"}).AddRow(1, "en")
	mock.ExpectExec("DELETE FROM userlang").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idlanguage, nameof\nFROM language")).WillReturnRows(rows)
	mock.ExpectExec("INSERT INTO userlang").WithArgs(int32(1), int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))

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
	runtimeconfig.AppRuntimeConfig.PageSizeDefault = 15
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
	ctx = context.WithValue(ctx, common.KeySession, sess)
	ctx = context.WithValue(ctx, common.KeyCoreData, &common.CoreData{})
	req = req.WithContext(ctx)

	prefRows := sqlmock.NewRows([]string{"idpreferences", "language_idlanguage", "users_idusers", "emailforumupdates", "page_size"}).
		AddRow(1, 1, 1, nil, runtimeconfig.AppRuntimeConfig.PageSizeDefault)
	mock.ExpectQuery("SELECT idpreferences").WithArgs(int32(1)).WillReturnRows(prefRows)
	mock.ExpectExec("UPDATE preferences").WithArgs(int32(2), int32(runtimeconfig.AppRuntimeConfig.PageSizeDefault), int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))

	userLangSaveLanguagePreferenceActionPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}
