package goa4web

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
	"github.com/arran4/goa4web/runtimeconfig"
	"github.com/gorilla/sessions"
)

func TestUserLangSaveAllActionPage_NewPref(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := New(db)
	store = sessions.NewCookieStore([]byte("test"))

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

	ctx := context.WithValue(req.Context(), ContextValues("queries"), queries)
	ctx = context.WithValue(ctx, ContextValues("session"), sess)
	ctx = context.WithValue(ctx, ContextValues("coreData"), &CoreData{})
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

	queries := New(db)
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

	ctx := context.WithValue(req.Context(), ContextValues("queries"), queries)
	ctx = context.WithValue(ctx, ContextValues("session"), sess)
	ctx = context.WithValue(ctx, ContextValues("coreData"), &CoreData{})
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

	queries := New(db)
	runtimeconfig.AppRuntimeConfig.PageSizeDefault = 15
	store = sessions.NewCookieStore([]byte("test"))

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

	ctx := context.WithValue(req.Context(), ContextValues("queries"), queries)
	ctx = context.WithValue(ctx, ContextValues("session"), sess)
	ctx = context.WithValue(ctx, ContextValues("coreData"), &CoreData{})
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
