package blogs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	db "github.com/arran4/goa4web/internal/db"
)

func TestBlogAddActionPage_InvalidForms(t *testing.T) {
	dbconn, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer dbconn.Close()

	queries := db.New(dbconn)
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"

	cases := []url.Values{
		{"text": {"hi"}},
		{"language": {"1"}},
		{"language": {"1"}, "text": {"hi"}, "foo": {"bar"}},
	}
	for _, form := range cases {
		req := httptest.NewRequest("POST", "/blogs/add", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		sess, _ := store.Get(req, core.SessionName)
		sess.Values["UID"] = int32(1)
		w := httptest.NewRecorder()
		sess.Save(req, w)
		for _, c := range w.Result().Cookies() {
			req.AddCookie(c)
		}
		ctx := context.WithValue(req.Context(), common.KeyQueries, queries)
		ctx = context.WithValue(ctx, common.KeyCoreData, &common.CoreData{})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		BlogAddActionPage(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("form=%v status=%d", form, rr.Code)
		}
	}
}
