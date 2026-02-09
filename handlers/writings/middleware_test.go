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

func TestRequireWritingView(t *testing.T) {
	// Setup session store
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"

	t.Run("Access Granted via View Grant", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/writings/article/1", nil)
		req = mux.SetURLVars(req, map[string]string{"writing": "1"})

		q := testhelpers.NewQuerierStub(
			testhelpers.WithGrant("writing", "article", "view"),
		)
		q.GetWritingForListerByIDRow = &db.GetWritingForListerByIDRow{Idwriting: 1, ForumthreadID: 100}

		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
		cd.SetCurrentSection("writing") // Required for SelectedThreadCanReply
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler := RequireWritingView(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		handler.ServeHTTP(rr, req)

		if status := rr.Result().StatusCode; status != http.StatusOK {
			t.Errorf("Expected 200, got status=%d", status)
		}
	})

	t.Run("Access Granted via Reply Grant (View Denied)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/writings/article/1", nil)
		req = mux.SetURLVars(req, map[string]string{"writing": "1"})

		q := testhelpers.NewQuerierStub()
		q.GetWritingForListerByIDRow = &db.GetWritingForListerByIDRow{Idwriting: 1, ForumthreadID: 100}

		// Override CheckGrant to see what's happening and allow reply
		q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
			if arg.Action == "reply" {
				return 1, nil
			}
			return 0, sql.ErrNoRows
		}

        // Explicitly stub thread query to ensure it returns error so HasGrant is called
		q.GetThreadBySectionThreadIDForReplierFn = func(ctx context.Context, arg db.GetThreadBySectionThreadIDForReplierParams) (*db.Forumthread, error) {
			return nil, sql.ErrNoRows
		}

		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
		cd.SetCurrentSection("writing") // Required for SelectedThreadCanReply
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler := RequireWritingView(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		handler.ServeHTTP(rr, req)

		if status := rr.Result().StatusCode; status != http.StatusOK {
			t.Errorf("Expected 200, got status=%d", status)
		}
	})

	t.Run("Access Denied", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/writings/article/1", nil)
		req = mux.SetURLVars(req, map[string]string{"writing": "1"})

		// No grants
		q := testhelpers.NewQuerierStub()
		q.GetWritingForListerByIDRow = &db.GetWritingForListerByIDRow{Idwriting: 1, ForumthreadID: 100}

		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
		cd.SetCurrentSection("writing") // Required for SelectedThreadCanReply
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler := RequireWritingView(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		handler.ServeHTTP(rr, req)

		if status := rr.Result().StatusCode; status != http.StatusForbidden {
			t.Errorf("Expected 403, got status=%d", status)
		}
	})

	t.Run("Writing Not Found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/writings/article/99", nil)
		req = mux.SetURLVars(req, map[string]string{"writing": "99"})

		q := testhelpers.NewQuerierStub()
		// No writing row set

		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
		// Section doesn't matter if writing not found
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler := RequireWritingView(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		handler.ServeHTTP(rr, req)

		// Expect 500 because handlers.RenderErrorPage(..., fmt.Errorf("No writing found")) defaults to 500
		if status := rr.Result().StatusCode; status != http.StatusInternalServerError {
			t.Errorf("Expected 500, got status=%d", status)
		}
	})
}
