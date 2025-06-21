package main

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
)

func TestLoginActionPageRedirect(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(login)).
		WithArgs(sql.NullString{String: "", Valid: true}, "").
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "passwd", "username"}).
			AddRow(1, "e", "p", ""))

	queries := New(db)

	store = sessions.NewCookieStore([]byte("test"))
	r := httptest.NewRequest("POST", "/login", nil)
	w := httptest.NewRecorder()
	sess, _ := store.Get(r, sessionName)
	sess.Values["return_path"] = "/dest"
	sess.Save(r, w)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}
	rr := httptest.NewRecorder()
	ctx := r.Context()
	ctx = context.WithValue(ctx, ContextValues("queries"), queries)
	ctx = context.WithValue(ctx, ContextValues("session"), sess)
	ctx = context.WithValue(ctx, ContextValues("coreData"), &CoreData{})
	r = r.WithContext(ctx)

	loginActionPage(rr, r)

	if rr.Result().StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
	if loc := rr.Result().Header.Get("Location"); loc != "/dest" {
		t.Errorf("loc=%s", loc)
	}
}

func TestLoginActionPageReturnForm(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(login)).
		WithArgs(sql.NullString{String: "", Valid: true}, "").
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "passwd", "username"}).
			AddRow(1, "e", "p", ""))

	queries := New(db)

	store = sessions.NewCookieStore([]byte("test"))
	r := httptest.NewRequest("POST", "/login", nil)
	w := httptest.NewRecorder()
	sess, _ := store.Get(r, sessionName)
	sess.Values["return_path"] = "/post"
	sess.Values["return_method"] = http.MethodPost
	vals := url.Values{"a": {"1"}}
	sess.Values["return_form"] = vals.Encode()
	sess.Save(r, w)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}
	rr := httptest.NewRecorder()
	ctx := r.Context()
	ctx = context.WithValue(ctx, ContextValues("queries"), queries)
	ctx = context.WithValue(ctx, ContextValues("session"), sess)
	ctx = context.WithValue(ctx, ContextValues("coreData"), &CoreData{})
	r = r.WithContext(ctx)

	loginActionPage(rr, r)

	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "name=\"a\" value=\"1\"") {
		t.Errorf("form data not present")
	}
}
