package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func TestBlogsBloggerPage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	q := New(db)
	store = sessions.NewCookieStore([]byte("test"))

	r := mux.NewRouter()
	br := r.PathPrefix("/blogs").Subrouter()
	br.HandleFunc("/blogger/{username}", blogsBloggerPage).Methods("GET")
	br.HandleFunc("/blogger/{username}/", blogsBloggerPage).Methods("GET")

	req := httptest.NewRequest("GET", "/blogs/blogger/bob", nil)

	sess, _ := store.Get(req, sessionName)
	sess.Values["UID"] = int32(1)
	w := httptest.NewRecorder()
	sess.Save(req, w)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}

	ctx := context.WithValue(req.Context(), ContextValues("queries"), q)
	ctx = context.WithValue(ctx, ContextValues("session"), sess)
	ctx = context.WithValue(ctx, ContextValues("coreData"), &CoreData{})
	req = req.WithContext(ctx)

       userRows := sqlmock.NewRows([]string{"idusers", "email", "passwd", "username"}).
               AddRow(1, "e", "p", "bob")
       mock.ExpectQuery(regexp.QuoteMeta(getUserByUsername)).
               WithArgs(sqlmock.AnyArg()).
               WillReturnRows(userRows)

	blogRows := sqlmock.NewRows([]string{
		"idblogs", "forumthread_idforumthread", "users_idusers",
		"language_idlanguage", "blog", "written", "username", "coalesce(th.comments, 0)",
	}).AddRow(1, 1, 1, 1, "hello", time.Unix(0, 0), "bob", 0)
       mock.ExpectQuery(regexp.QuoteMeta("SELECT b.idblogs")).
               WithArgs(int32(1), int32(1), int32(1), int32(1), int32(15), int32(0)).
               WillReturnRows(blogRows)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
