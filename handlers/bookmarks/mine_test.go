package bookmarks

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

type fakeBookmarksQuerier struct {
	db.Querier
	row *db.GetBookmarksForUserRow
	err error
}

func (f fakeBookmarksQuerier) GetBookmarksForUser(ctx context.Context, usersID int32) (*db.GetBookmarksForUserRow, error) {
	return f.row, f.err
}

func (f fakeBookmarksQuerier) GetPermissionsByUserID(ctx context.Context, usersID int32) ([]*db.GetPermissionsByUserIDRow, error) {
	return nil, nil
}

func (f fakeBookmarksQuerier) SystemCheckRoleGrant(ctx context.Context, arg db.SystemCheckRoleGrantParams) (int32, error) {
	return 0, errors.New("not granted")
}

func newBookmarksRequest(t *testing.T, q db.Querier) *http.Request {
	t.Helper()

	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"

	req := httptest.NewRequest("GET", "/bookmarks/mine", nil)
	sess, err := store.Get(req, core.SessionName)
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	sess.Values["UID"] = int32(1)

	w := httptest.NewRecorder()
	if err := sess.Save(req, w); err != nil {
		t.Fatalf("save session: %v", err)
	}
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}

	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithSession(sess))
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	return req.WithContext(ctx)
}

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

func TestMinePage_NoBookmarks(t *testing.T) {
	req := newBookmarksRequest(t, fakeBookmarksQuerier{err: sql.ErrNoRows})

	rr := httptest.NewRecorder()
	MinePage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "No bookmarks saved") {
		t.Fatalf("body=%q", rr.Body.String())
	}
}

func TestMinePage_RenderBookmarks(t *testing.T) {
	bookmarks := &db.GetBookmarksForUserRow{
		List: sql.NullString{
			String: "Category: Reference\nhttps://example.com Example\nCategory: Search\nhttps://search.example Search\n",
			Valid:  true,
		},
	}
	req := newBookmarksRequest(t, fakeBookmarksQuerier{row: bookmarks})

	rr := httptest.NewRecorder()
	MinePage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()
	for _, want := range []string{
		"<h2>Reference</h2>",
		`<a href="https://example.com" target="_blank">Example</a>`,
		"<h2>Search</h2>",
		`<a href="https://search.example" target="_blank">Search</a>`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected %q in body=%q", want, body)
		}
	}
}
