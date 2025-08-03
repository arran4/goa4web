package bookmarks

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
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

func TestMinePage_NoBookmarks(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT Idbookmarks, list\nFROM bookmarks\nWHERE users_idusers = ?")).
		WithArgs(int32(1)).WillReturnError(sql.ErrNoRows)

	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"

	req := httptest.NewRequest("GET", "/bookmarks/mine", nil)
	sess, _ := store.Get(req, core.SessionName)
	sess.Values["UID"] = int32(1)
	w := httptest.NewRecorder()
	sess.Save(req, w)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}

	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithSession(sess))
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	MinePage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if !strings.Contains(rr.Body.String(), "No bookmarks saved") {
		t.Fatalf("body=%q", rr.Body.String())
	}
}
