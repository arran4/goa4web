package forum

import (
	"github.com/arran4/goa4web/handlers/forumcommon"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestRequireThreadAndTopicTrue(t *testing.T) {
	threadID := int32(2)
	topicID := int32(1)

	qs := testhelpers.NewQuerierStub()
	qs.GetThreadLastPosterAndPermsFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
		if arg.ThreadID != threadID {
			return nil, sql.ErrNoRows
		}
		return &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:          threadID,
			ForumtopicIdforumtopic: topicID,
		}, nil
	}
	qs.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		if arg.Idforumtopic != topicID {
			return nil, sql.ErrNoRows
		}
		return &db.GetForumTopicByIdForUserRow{
			Idforumtopic: topicID,
		}, nil
	}

	req := httptest.NewRequest("GET", "/forum/topic/1/thread/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "2"})
	cd := common.NewCoreData(req.Context(), qs, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	called := false
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if _, err := cd.SelectedThread(); err != nil {
			t.Errorf("SelectedThread: %v", err)
		}
		if _, err := cd.CurrentTopic(); err != nil {
			t.Errorf("CurrentTopic: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	})

	forumcommon.RequireThreadAndTopic(handler).ServeHTTP(rr, req)
	if !called {
		t.Errorf("expected handler call")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("unexpected status %d", rr.Code)
	}
}

func TestRequireThreadAndTopicFalse(t *testing.T) {
	threadID := int32(2)
	wrongTopicID := int32(3)

	qs := testhelpers.NewQuerierStub()
	qs.GetThreadLastPosterAndPermsFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
		return &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:          threadID,
			ForumtopicIdforumtopic: wrongTopicID,
		}, nil
	}
	qs.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		return &db.GetForumTopicByIdForUserRow{
			Idforumtopic: wrongTopicID,
		}, nil
	}

	req := httptest.NewRequest("GET", "/forum/topic/1/thread/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "2"})
	cd := common.NewCoreData(req.Context(), qs, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	called := false
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	forumcommon.RequireThreadAndTopic(handler).ServeHTTP(rr, req)
	if called {
		t.Errorf("expected handler not called")
	}
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 got %d", rr.Code)
	}
}

func TestRequireThreadAndTopicError(t *testing.T) {
	qs := testhelpers.NewQuerierStub()
	qs.GetThreadLastPosterAndPermsFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
		return nil, sql.ErrNoRows
	}
	qs.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		return nil, sql.ErrNoRows
	}

	req := httptest.NewRequest("GET", "/forum/topic/1/thread/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "2"})
	cd := common.NewCoreData(req.Context(), qs, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	called := false
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	forumcommon.RequireThreadAndTopic(handler).ServeHTTP(rr, req)
	if called {
		t.Errorf("expected handler not called")
	}
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 got %d", rr.Code)
	}
}
