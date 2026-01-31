package writings

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
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestRequireWritingAuthorWritingVar(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"

	req := httptest.NewRequest("GET", "/writings/article/2/edit", nil)
	req = mux.SetURLVars(req, map[string]string{"writing": "2"})

	sess := testhelpers.Must(store.Get(req, core.SessionName))
	sess.Values["UID"] = int32(1)

	cd := common.NewCoreData(
		req.Context(),
		q,
		config.NewRuntimeConfig(),
		common.WithSession(sess),
		common.WithUserRoles([]string{"content writer"}),
	)
	cd.LoadSelectionsFromRequest(req)
	cd.UserID = 1
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	q.GetWritingForListerByIDRow = &db.GetWritingForListerByIDRow{
		Idwriting:         2,
		UsersIdusers:      1,
		ForumthreadID:     0,
		LanguageID:        sql.NullInt32{Int32: 1, Valid: true},
		WritingCategoryID: 1,
		Title:             sql.NullString{},
		Published:         sql.NullTime{},
		Timezone:          sql.NullString{},
		Writing:           sql.NullString{},
		Abstract:          sql.NullString{},
		Private:           sql.NullBool{},
		DeletedAt:         sql.NullTime{},
		LastIndex:         sql.NullTime{},
		Writerid:          1,
		Writerusername:    sql.NullString{},
	}
	q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		return 1, nil
	}

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if cd.CurrentWritingLoaded() == nil {
			t.Errorf("writing not cached")
		}
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	RequireWritingAuthor(handler).ServeHTTP(rr, req)

	if !called {
		t.Errorf("expected handler call")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status=%d", rr.Code)
	}
}

func TestMatchCanEditWritingArticleUsesGrant(t *testing.T) {
	req := httptest.NewRequest("GET", "/writings/article/2/edit", nil)
	req = mux.SetURLVars(req, map[string]string{"writing": "2"})

	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		return 1, nil
	}
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	cd.UserID = 7
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if !MatchCanEditWritingArticle(req, &mux.RouteMatch{}) {
		t.Fatalf("expected match to allow edit")
	}
}

func TestMatchCanPostWritingDeniesWithoutGrant(t *testing.T) {
	req := httptest.NewRequest("GET", "/writings/category/3/add", nil)
	req = mux.SetURLVars(req, map[string]string{"category": "3"})

	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		return 0, sql.ErrNoRows
	}
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	cd.UserID = 4
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if MatchCanPostWriting(req, &mux.RouteMatch{}) {
		t.Fatalf("expected match to deny posting")
	}
}
