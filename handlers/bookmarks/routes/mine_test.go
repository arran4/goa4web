package routes

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers/handlertest"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

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
