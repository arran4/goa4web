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

	"github.com/arran4/goa4web/config"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestBlogsBloggerPostsPage(t *testing.T) {
	r := mux.NewRouter()
	br := r.PathPrefix("/blogs").Subrouter()
	br.HandleFunc("/blogger/{username}", BloggerPostsPage).Methods("GET")
	br.HandleFunc("/blogger/{username}/", BloggerPostsPage).Methods("GET")

	req := httptest.NewRequest("GET", "/blogs/blogger/bob", nil)
	rr := httptest.NewRecorder()
	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	r.ServeHTTP(rr, req)

	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}

func TestBlogsRssPageWritesRSS(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT idusers, (SELECT email FROM user_emails ue WHERE ue.user_id = users.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email, username, public_profile_enabled_at\nFROM users\nWHERE username = ?")).
		WithArgs("bob").
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).
			AddRow(1, "e", "bob", nil))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT b.idblogs")).
		WithArgs(int32(1), int32(1), int32(1), int32(1), int32(1), int32(1), sqlmock.AnyArg(), int32(15), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"idblogs", "forumthread_idforumthread", "users_idusers", "language_idlanguage", "blog", "written", "username", "coalesce(th.comments, 0)", "is_owner"}).
			AddRow(1, 1, 1, 1, "hello", time.Unix(0, 0), "bob", 0, true))

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
	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
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
	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
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
	req := httptest.NewRequest("GET", "/admin/blogs/users/roles", nil)
	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
	cd.SetRoles([]string{"anonymous"})
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	GetPermissionsByUserIdAndSectionBlogsPage(rr, req)
	if rr.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Result().StatusCode)
	}
}
