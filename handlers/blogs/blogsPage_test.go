package blogs

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/consts"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/sharesign"
)

var (
	store       *sessions.CookieStore
	sessionName = "test-session"
)

func TestBlogsBloggerPostsPage(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)
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
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithSession(sess))
	cd.ShareSigner = sharesign.NewSigner(config.NewRuntimeConfig(), "secret")
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	userRows := sqlmock.NewRows([]string{"idusers", "username", "public_profile_enabled_at"}).
		AddRow(1, "bob", nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idusers, username, public_profile_enabled_at\nFROM users\nWHERE username = ?")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(userRows)

	blogRows := sqlmock.NewRows([]string{
		"idblogs", "forumthread_id", "users_idusers",
		"language_id", "blog", "written", "timezone", "username", "coalesce(th.comments, 0)", "is_owner",
	}).AddRow(1, 1, 1, 1, "hello", time.Unix(0, 0), time.Local.String(), "bob", 0, true)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT b.idblogs")).
		WithArgs(int32(1), int32(1), int32(1), int32(1), int32(1), int32(1), sqlmock.AnyArg(), int32(15), int32(0)).
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
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT idusers, username, public_profile_enabled_at\nFROM users\nWHERE username = ?")).
		WithArgs("bob").
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "username", "public_profile_enabled_at"}).
			AddRow(1, "bob", nil))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT b.idblogs")).
		WithArgs(int32(0), int32(0), int32(1), int32(1), int32(0), int32(0), sqlmock.AnyArg(), int32(15), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"idblogs", "forumthread_id", "users_idusers", "language_id", "blog", "written", "timezone", "username", "coalesce(th.comments, 0)", "is_owner", "title"}).
			AddRow(1, 1, 1, 1, "hello", time.Unix(0, 0), time.Local.String(), "bob", 0, true, "hello"))

	req := httptest.NewRequest("GET", "http://example.com/blogs/rss?rss=bob", nil)
	q := db.New(conn)
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
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
	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
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
	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	BlogEditPage(rr, req)
	if rr.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
	}
}
