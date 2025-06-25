package goa4web

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/handlers/common"
)

func TestBlogsBloggerPage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	q := New(db)
	store = sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = sessionName

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

	ctx := context.WithValue(req.Context(), common.KeyQueries, q)
	ctx = context.WithValue(ctx, common.KeySession, sess)
	ctx = context.WithValue(ctx, common.KeyCoreData, &CoreData{})
	req = req.WithContext(ctx)

	userRows := sqlmock.NewRows([]string{"idusers", "email", "passwd", "passwd_algorithm", "username"}).
		AddRow(1, "e", "p", "", "bob")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idusers, email, passwd, passwd_algorithm, username\nFROM users\nWHERE username = ?")).
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

func TestBlogsRssPageWritesRSS(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := New(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT idusers, email, passwd, passwd_algorithm, username\nFROM users\nWHERE username = ?")).
		WithArgs("bob").
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "passwd", "passwd_algorithm", "username"}).
			AddRow(1, "e", "p", "", "bob"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT b.idblogs")).
		WithArgs(int32(1), int32(1), int32(1), int32(1), int32(15), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"idblogs", "forumthread_idforumthread", "users_idusers", "language_idlanguage", "blog", "written", "username", "coalesce(th.comments, 0)"}).
			AddRow(1, 1, 1, 1, "hello", time.Unix(0, 0), "bob", 0))

	req := httptest.NewRequest("GET", "http://example.com/blogs/rss?rss=bob", nil)
	ctx := context.WithValue(req.Context(), common.KeyQueries, queries)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	blogsRssPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/rss+xml" {
		t.Errorf("Content-Type=%q", ct)
	}

	var v struct{ XMLName xml.Name }
	if err := xml.Unmarshal(rr.Body.Bytes(), &v); err != nil {
		t.Fatalf("xml parse: %v", err)
	}
	if v.XMLName.Local != "rss" {
		t.Errorf("expected root rss got %s", v.XMLName.Local)
	}
}

func TestBlogsBlogAddPage_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	ctx := context.WithValue(req.Context(), common.KeyCoreData, &CoreData{SecurityLevel: "reader"})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	blogsBlogAddPage(rr, req)
	if rr.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
	}
}

func TestBlogsBlogEditPage_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/1/edit", nil)
	ctx := context.WithValue(req.Context(), common.KeyCoreData, &CoreData{SecurityLevel: "reader"})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	blogsBlogEditPage(rr, req)
	if rr.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
	}
}

func TestGetPermissionsByUserIdAndSectionBlogsPage_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "/admin/blogs/user/permissions", nil)
	ctx := context.WithValue(req.Context(), common.KeyCoreData, &CoreData{SecurityLevel: "reader"})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	getPermissionsByUserIdAndSectionBlogsPage(rr, req)
	if rr.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
	}
}
