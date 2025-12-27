package news

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestRequireNewsPostAuthor_AllowsOwner(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/news/news/5/edit", nil)
	req = mux.SetURLVars(req, map[string]string{"news": "5"})

	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"

	sess, _ := store.Get(req, core.SessionName)
	sess.Values["UID"] = int32(7)
	w := httptest.NewRecorder()
	if err := sess.Save(req, w); err != nil {
		t.Fatalf("save session: %v", err)
	}
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}

	q := &db.QuerierStub{
		GetForumThreadIdByNewsPostIdRow: &db.GetForumThreadIdByNewsPostIdRow{
			ForumthreadID: 21,
			Idusers:       sql.NullInt32{Int32: 7, Valid: true},
		},
	}
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	var called bool
	rr := httptest.NewRecorder()
	RequireNewsPostAuthor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if cd.CurrentNewsPostLoaded() == nil {
			t.Fatalf("expected current news post to be cached")
		}
		w.WriteHeader(http.StatusTeapot)
	})).ServeHTTP(rr, req)

	if !called {
		t.Fatalf("expected next handler to be called")
	}
	if rr.Code != http.StatusTeapot {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if len(q.GetForumThreadIdByNewsPostIdCalls) != 1 || q.GetForumThreadIdByNewsPostIdCalls[0] != 5 {
		t.Fatalf("unexpected call log: %+v", q.GetForumThreadIdByNewsPostIdCalls)
	}
	if cd.CurrentNewsPostLoaded().ForumthreadID != 21 {
		t.Fatalf("unexpected cached post: %+v", cd.CurrentNewsPostLoaded())
	}
}

func TestRequireNewsPostAuthor_RejectsNonOwner(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/news/news/5/edit", nil)
	req = mux.SetURLVars(req, map[string]string{"news": "5"})

	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"

	sess, _ := store.Get(req, core.SessionName)
	sess.Values["UID"] = int32(8)
	w := httptest.NewRecorder()
	if err := sess.Save(req, w); err != nil {
		t.Fatalf("save session: %v", err)
	}
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}

	q := &db.QuerierStub{
		GetForumThreadIdByNewsPostIdRow: &db.GetForumThreadIdByNewsPostIdRow{
			ForumthreadID: 21,
			Idusers:       sql.NullInt32{Int32: 7, Valid: true},
		},
	}
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	RequireNewsPostAuthor(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatalf("handler should not be called")
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if len(q.GetForumThreadIdByNewsPostIdCalls) != 1 {
		t.Fatalf("expected one retrieval call, got %d", len(q.GetForumThreadIdByNewsPostIdCalls))
	}
	if cd.CurrentNewsPostLoaded() != nil {
		t.Fatalf("expected no cached post on rejection, got %+v", cd.CurrentNewsPostLoaded())
	}
}
