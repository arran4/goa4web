package goa4web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/sessions"
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
	req, rr := newRequestWithSession("GET", "/", map[string]interface{}{
		"UID":        int32(1),
		"ExpiryTime": time.Now().Add(-time.Hour).Unix(),
	})
	ctx := context.WithValue(req.Context(), ContextValues("queries"), New(nil))
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

	queries := New(db)
	ctx := context.WithValue(req.Context(), ContextValues("queries"), queries)
	req = req.WithContext(ctx)

	mock.ExpectQuery("SELECT idusers, email, passwd, username").
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "passwd", "username"}).
			AddRow(1, "e", "p", "u"))
	mock.ExpectQuery("SELECT idpermissions, users_idusers, section, level FROM permissions WHERE users_idusers = ?").
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"idpermissions", "users_idusers", "section", "level"}).AddRow(1, 1, "all", "admin"))
	mock.ExpectQuery("SELECT idpreferences, language_idlanguage, users_idusers, emailforumupdates, page_size FROM preferences WHERE users_idusers = ?").
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"idpreferences", "language_idlanguage", "users_idusers", "emailforumupdates", "page_size"}).AddRow(1, 2, 1, false, 15))
	mock.ExpectQuery("SELECT iduserlang, users_idusers, language_idlanguage FROM userlang WHERE users_idusers = ?").
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"iduserlang", "users_idusers", "language_idlanguage"}).AddRow(1, 1, 2))

	var gotPerms []*Permission
	var gotPref *Preference
	var gotLangs []*Userlang

	handler := UserAdderMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPerms, _ = r.Context().Value(ContextValues("permissions")).([]*Permission)
		gotPref, _ = r.Context().Value(ContextValues("preference")).(*Preference)
		gotLangs, _ = r.Context().Value(ContextValues("languages")).([]*Userlang)
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
