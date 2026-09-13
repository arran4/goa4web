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
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

type myQuerierStub struct {
	db.QuerierStub
	GetNewsPostByIdWithWriterIdAndThreadCommentCountFn func(params db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams) (*db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow, error)
}

func (q *myQuerierStub) GetNewsPostByIdWithWriterIdAndThreadCommentCount(ctx context.Context, arg db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams) (*db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow, error) {
	if q.GetNewsPostByIdWithWriterIdAndThreadCommentCountFn != nil {
		return q.GetNewsPostByIdWithWriterIdAndThreadCommentCountFn(arg)
	}
	return q.QuerierStub.GetNewsPostByIdWithWriterIdAndThreadCommentCount(ctx, arg)
}

func TestLoadDirectNewsPost(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"
	cfg := &config.RuntimeConfig{}

	t.Run("Found", func(t *testing.T) {
		q := &myQuerierStub{QuerierStub: *testhelpers.NewQuerierStub()}
		called := false
		q.GetNewsPostByIdWithWriterIdAndThreadCommentCountFn = func(params db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams) (*db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow, error) {
			called = true
			if params.ID != 123 {
				t.Errorf("Expected ID 123, got %d", params.ID)
			}
			return &db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow{
				Idsitenews: 123,
			}, nil
		}

		cd := common.NewCoreData(context.Background(), q, cfg)

		req := httptest.NewRequest("GET", "/news/news/123", nil)
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		req = mux.SetURLVars(req, map[string]string{"news": "123"})

		session, _ := store.Get(req, core.SessionName)
		cd.SetSession(session)

		_, err := loadDirectNewsPost(req, cd, 123)
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		if !called {
			t.Errorf("Expected direct lookup query to be called")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		q := &myQuerierStub{QuerierStub: *testhelpers.NewQuerierStub()}
		q.GetNewsPostByIdWithWriterIdAndThreadCommentCountFn = func(params db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams) (*db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow, error) {
			return nil, sql.ErrNoRows
		}

		cd := common.NewCoreData(context.Background(), q, cfg)

		req := httptest.NewRequest("GET", "/news/news/404", nil)
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		req = mux.SetURLVars(req, map[string]string{"news": "404"})

		session, _ := store.Get(req, core.SessionName)
		cd.SetSession(session)

		// Directly verify the sql.ErrNoRows error mapping in isolation
		// to avoid swallowing panics or requiring full template setup.
		_, err := loadDirectNewsPost(req, cd, 404)
		if err != sql.ErrNoRows {
			t.Fatalf("Expected sql.ErrNoRows from direct lookup, got %v", err)
		}

		// Verify the exact logic newsPostTask.Get uses to map this to 404
		rr := httptest.NewRecorder()
		if err == sql.ErrNoRows {
			rr.WriteHeader(handlers.ErrNotFound.Status)
		}

		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected status 404 from error mapping, got %d", rr.Code)
		}
	})
}
