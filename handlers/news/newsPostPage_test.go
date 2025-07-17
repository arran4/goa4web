package news

import (
	"context"
	corecommon "github.com/arran4/goa4web/core/common"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

func TestNewsPostNewActionPage_InvalidForms(t *testing.T) {
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
		req := httptest.NewRequest("POST", "/news", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		sess, _ := store.Get(req, core.SessionName)
		sess.Values["UID"] = int32(1)
		w := httptest.NewRecorder()
		sess.Save(req, w)
		for _, c := range w.Result().Cookies() {
			req.AddCookie(c)
		}
		ctx := context.WithValue(req.Context(), handlers.KeyQueries, queries)
		ctx = context.WithValue(ctx, handlers.KeyCoreData, &corecommon.CoreData{})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		newPostTask.Action(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("form=%v status=%d", form, rr.Code)
		}
		if req.URL.RawQuery == "" {
			t.Errorf("query not set")
		}
		if !strings.Contains(rr.Body.String(), "<a href=") {
			t.Errorf("body=%q", rr.Body.String())
		}
	}
}

func TestNewsPostEditActionPage_InvalidForms(t *testing.T) {
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
		req := httptest.NewRequest("POST", "/news/1", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = mux.SetURLVars(req, map[string]string{"post": "1"})
		sess, _ := store.Get(req, core.SessionName)
		sess.Values["UID"] = int32(1)
		w := httptest.NewRecorder()
		sess.Save(req, w)
		for _, c := range w.Result().Cookies() {
			req.AddCookie(c)
		}
		ctx := context.WithValue(req.Context(), handlers.KeyQueries, queries)
		ctx = context.WithValue(ctx, handlers.KeyCoreData, &corecommon.CoreData{})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		editTask.Action(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("form=%v status=%d", form, rr.Code)
		}
		if req.URL.RawQuery == "" {
			t.Errorf("query not set")
		}
		if !strings.Contains(rr.Body.String(), "<a href=") {
			t.Errorf("body=%q", rr.Body.String())
		}
	}
}
