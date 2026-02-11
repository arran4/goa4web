package writings

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestArticleReplyActionPage(t *testing.T) {
	t.Run("Happy Path - Uses Writing Param", func(t *testing.T) {
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"

		form := url.Values{}
		form.Set("replytext", "hi")
		form.Set("language", "invalid")
		req := httptest.NewRequest("POST", "/writings/article/1", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = mux.SetURLVars(req, map[string]string{"writing": "1"})

		w := httptest.NewRecorder()
		sess := testhelpers.Must(store.Get(req, core.SessionName))
		sess.Values["UID"] = int32(1)
		sess.Save(req, w)
		for _, c := range w.Result().Cookies() {
			req.AddCookie(c)
		}

		q := testhelpers.NewQuerierStub(testhelpers.WithGrantResult(true))
		q.GetWritingForListerByIDRow = &db.GetWritingForListerByIDRow{Idwriting: 1}
		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithSession(sess), common.WithUserRoles([]string{"user"}))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(replyTask)(rr, req)

		if rr.Result().StatusCode != http.StatusInternalServerError {
			t.Fatalf("status=%d", rr.Result().StatusCode)
		}

		t.Run("Data Consequences", func(t *testing.T) {
			if got := len(q.GetWritingForListerByIDCalls); got != 1 {
				t.Fatalf("GetWritingForListerByIDCalls=%d", got)
			}
			call := q.GetWritingForListerByIDCalls[0]
			if call.Idwriting != 1 || call.ListerID != 1 || call.ListerMatchID != (sql.NullInt32{Int32: 1, Valid: true}) {
				t.Fatalf("unexpected article lookup: %+v", call)
			}
			if got := len(q.SystemCheckGrantCalls); got != 1 {
				t.Fatalf("SystemCheckGrantCalls=%d", got)
			}
			grant := q.SystemCheckGrantCalls[0]
			if grant.Section != "writing" || grant.Item.String != "article" || grant.Action != "reply" || grant.ItemID != (sql.NullInt32{Int32: 1, Valid: true}) {
				t.Fatalf("unexpected grant check: %+v", grant)
			}
		})
	})
}
