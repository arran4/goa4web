package blogs

import (
	"context"
	"database/sql"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/consts"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

var (
	store       *sessions.CookieStore
	sessionName = "test-session"
)

func TestBlogsBloggerPostsPage(t *testing.T) {
	q := testhelpers.NewQuerierStub(testhelpers.WithGrant("blogs", "entry", "see"))
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
	cd.ShareSignKey = "secret"
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	q.SystemGetUserByUsernameRow = &db.SystemGetUserByUsernameRow{
		Idusers:                1,
		Username:               sql.NullString{String: "bob", Valid: true},
		PublicProfileEnabledAt: sql.NullTime{},
	}
	q.ListBlogEntriesByAuthorForListerReturns = []*db.ListBlogEntriesByAuthorForListerRow{
		{
			Idblogs:       1,
			ForumthreadID: sql.NullInt32{Int32: 1, Valid: true},
			UsersIdusers:  1,
			LanguageID:    sql.NullInt32{Int32: 1, Valid: true},
			Blog:          sql.NullString{String: "hello", Valid: true},
			Written:       time.Unix(0, 0),
			Timezone:      sql.NullString{String: time.Local.String(), Valid: true},
			Username:      sql.NullString{String: "bob", Valid: true},
			Comments:      0,
			IsOwner:       true,
			Title:         "hello",
		},
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}

func TestBlogsRssPageWritesRSS(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.SystemGetUserByUsernameRow = &db.SystemGetUserByUsernameRow{
		Idusers:                1,
		Username:               sql.NullString{String: "bob", Valid: true},
		PublicProfileEnabledAt: sql.NullTime{},
	}
	q.ListBlogEntriesByAuthorForListerReturns = []*db.ListBlogEntriesByAuthorForListerRow{
		{
			Idblogs:       1,
			ForumthreadID: sql.NullInt32{Int32: 1, Valid: true},
			UsersIdusers:  1,
			LanguageID:    sql.NullInt32{Int32: 1, Valid: true},
			Blog:          sql.NullString{String: "hello", Valid: true},
			Written:       time.Unix(0, 0),
			Timezone:      sql.NullString{String: time.Local.String(), Valid: true},
			Username:      sql.NullString{String: "bob", Valid: true},
			Comments:      0,
			IsOwner:       true,
			Title:         "hello",
		},
	}

	req := httptest.NewRequest("GET", "http://example.com/blogs/rss?rss=bob", nil)
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithSiteTitle("Site"))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	RssPage(rr, req)

	if ct := rr.Header().Get("Content-Type"); ct != "application/rss+xml" {
		t.Errorf("Content-Type=%q", ct)
	}

	var v struct {
		XMLName xml.Name
		Channel struct {
			Title string `xml:"title"`
		} `xml:"channel"`
	}
	if err := xml.Unmarshal(rr.Body.Bytes(), &v); err != nil {
		t.Fatalf("xml parse: %v", err)
	}
	if v.XMLName.Local != "rss" {
		t.Errorf("expected root rss got %s", v.XMLName.Local)
	}
	if v.Channel.Title != "Site - bob blog" {
		t.Errorf("expected title 'Site - bob blog' got %q", v.Channel.Title)
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
