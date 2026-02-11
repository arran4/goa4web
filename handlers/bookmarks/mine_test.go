package bookmarks

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers/handlertest"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestParseColumns(t *testing.T) {
	tests := []struct {
		name      string
		bookmarks string
		want      []*Column
	}{
		{
			name:      "Test",
			bookmarks: "Category: Search\nhttp://www.google.com.au Google\nCategory: Wikies\nhttp://en.wikipedia.org/wiki/Main_Page Wikipedia\nhttp://mathworld.wolfram.com/ Math World\nhttp://gentoo-wiki.com/Main_Page Gentoo-wiki\n",
			want: []*Column{{
				Categories: []*Category{
					{
						Name: "Search",
						Entries: []*Entry{
							{
								Url:  "http://www.google.com.au",
								Name: "Google",
							},
						},
					},
					{
						Name: "Wikies",
						Entries: []*Entry{
							{
								Url:  "http://en.wikipedia.org/wiki/Main_Page",
								Name: "Wikipedia",
							},
							{
								Url:  "http://mathworld.wolfram.com/",
								Name: "Math World",
							},
							{
								Url:  "http://gentoo-wiki.com/Main_Page",
								Name: "Gentoo-wiki",
							},
						},
					},
				}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseColumns(tt.bookmarks)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("ParseColumns() = diff\n%s", diff)
			}
		})
	}
}

func TestMinePage(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/bookmarks/mine", nil)
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"

		sess := testhelpers.Must(store.Get(req, core.SessionName))
		sess.Values["UID"] = int32(1)
		w := httptest.NewRecorder()
		sess.Save(req, w)
		for _, c := range w.Result().Cookies() {
			req.AddCookie(c)
		}

		req, cd, stub := handlertest.RequestWithCoreData(t, req, common.WithSession(sess))
		cd.UserID = 1

		stub.GetBookmarksForUserReturns = &db.GetBookmarksForUserRow{
			Idbookmarks: 1,
			List: sql.NullString{
				String: "Category: Test\nhttp://example.com Example",
				Valid:  true,
			},
		}

		rr := httptest.NewRecorder()
		MinePage(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Example") {
			t.Fatalf("body=%q expected %q", rr.Body.String(), "Example")
		}
		if cd.PageTitle != "My Bookmarks" {
			t.Fatalf("PageTitle=%q expected %q", cd.PageTitle, "My Bookmarks")
		}
	})

	t.Run("No Bookmarks", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/bookmarks/mine", nil)
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"

		sess := testhelpers.Must(store.Get(req, core.SessionName))
		sess.Values["UID"] = int32(1)
		w := httptest.NewRecorder()
		sess.Save(req, w)
		for _, c := range w.Result().Cookies() {
			req.AddCookie(c)
		}

		req, cd, stub := handlertest.RequestWithCoreData(t, req, common.WithSession(sess))
		cd.UserID = 1

		stub.GetBookmarksForUserErr = sql.ErrNoRows

		rr := httptest.NewRecorder()
		MinePage(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "No bookmarks saved") {
			t.Fatalf("body=%q", rr.Body.String())
		}
	})
}
