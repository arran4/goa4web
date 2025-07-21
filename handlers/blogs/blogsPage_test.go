package blogs

import (
	"context"
	"encoding/xml"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	db "github.com/arran4/goa4web/internal/db"
)

var (
	store       *sessions.CookieStore
	sessionName = "test-session"
)

func TestBlogsBloggerPostsPage(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := db.New(sqldb)
	store = sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = sessionName

	r := mux.NewRouter()
	br := r.PathPrefix("/blogs").Subrouter()
	br.HandleFunc("/blogger/{username}", BloggerPostsPage).Methods("GET")
	br.HandleFunc("/blogger/{username}/", BloggerPostsPage).Methods("GET")

	req := httptest.NewRequest("GET", "/blogs/blogger/bob", nil)

	sess, _ := store.Get(req, sessionName)
	sess.Values["UID"] = int32(1)
	w := httptest.NewRecorder()
	sess.Save(req, w)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}

	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, common.WithSession(sess))
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	userRows := sqlmock.NewRows([]string{"idusers", "email", "username"}).
		AddRow(1, "e", "bob")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idusers, (SELECT email FROM user_emails ue WHERE ue.user_id = users.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email, username\nFROM users\nWHERE username = ?")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(userRows)

	blogRows := sqlmock.NewRows([]string{
		"idblogs", "forumthread_idforumthread", "users_idusers",
		"language_idlanguage", "blog", "written", "username", "coalesce(th.comments, 0)", "is_owner",
	}).AddRow(1, 1, 1, 1, "hello", time.Unix(0, 0), "bob", 0, true)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT b.idblogs")).
		WithArgs(int32(1), int32(1), int32(1), int32(1), int32(1), int32(15), int32(0)).
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
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	queries := db.New(sqldb)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT idusers, (SELECT email FROM user_emails ue WHERE ue.user_id = users.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email, username\nFROM users\nWHERE username = ?")).
		WithArgs("bob").
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).
			AddRow(1, "e", "bob"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT b.idblogs")).
		WithArgs(int32(1), int32(1), int32(1), int32(1), int32(1), int32(15), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"idblogs", "forumthread_idforumthread", "users_idusers", "language_idlanguage", "blog", "written", "username", "coalesce(th.comments, 0)", "is_owner"}).
			AddRow(1, 1, 1, 1, "hello", time.Unix(0, 0), "bob", 0, true))

	req := httptest.NewRequest("GET", "http://example.com/blogs/rss?rss=bob", nil)
	ctx := req.Context()
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	RssPage(rr, req)

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
	cd := common.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"anonymous"})
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	BlogAddPage(rr, req)
	if rr.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
	}
}

func TestBlogsBlogEditPage_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/1/edit", nil)
	cd := common.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"anonymous"})
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	BlogEditPage(rr, req)
	if rr.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
	}
}

func TestGetPermissionsByUserIdAndSectionBlogsPage_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "/admin/blogs/user/permissions", nil)
	cd := common.NewCoreData(req.Context(), nil)
	cd.SetRoles([]string{"anonymous"})
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	GetPermissionsByUserIdAndSectionBlogsPage(rr, req)
	if rr.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
	}
}
