package news

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
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

func TestNewsPostTask_Get_DirectLookup(t *testing.T) {
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

		cd := common.NewCoreData(context.Background(), q, nil)
		session := sessions.NewSession(sessions.NewCookieStore([]byte("secret")), "test-session")
		cd.SetSession(session)

		req := httptest.NewRequest("GET", "/news/news/123", nil)
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		req = mux.SetURLVars(req, map[string]string{"news": "123"})

		rr := httptest.NewRecorder()

		task := &newsPostTask{}

		defer func() {
			recover() // Ignore template panics
			if !called {
				t.Errorf("Expected direct lookup query to be called")
			}
		}()

		task.Get(rr, req)
	})

	t.Run("NotFound", func(t *testing.T) {
		q := &myQuerierStub{QuerierStub: *testhelpers.NewQuerierStub()}
		q.GetNewsPostByIdWithWriterIdAndThreadCommentCountFn = func(params db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams) (*db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow, error) {
			return nil, sql.ErrNoRows
		}

		cd := common.NewCoreData(context.Background(), q, nil)
		session := sessions.NewSession(sessions.NewCookieStore([]byte("secret")), "test-session")
		cd.SetSession(session)

		req := httptest.NewRequest("GET", "/news/news/404", nil)

		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		req = mux.SetURLVars(req, map[string]string{"news": "404"})

		rr := httptest.NewRecorder()

		task := &newsPostTask{}

		defer func() {
			recover()
			if rr.Code != http.StatusNotFound {
				t.Errorf("Expected status 404, got %d", rr.Code)
			}
		}()

		task.Get(rr, req)
	})
}
