package privateforum

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
)

func TestUnreadPrivateThreadsFiltering(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	var capturedTopicIDNull interface{}
	var capturedTopicIDVal int32

	q.ListUnreadPrivateThreadsForUserFn = func(ctx context.Context, arg db.ListUnreadPrivateThreadsForUserParams) ([]*db.ListUnreadPrivateThreadsForUserRow, error) {
		capturedTopicIDNull = arg.TopicIDNull
		capturedTopicIDVal = arg.TopicIDVal
		return nil, nil
	}
	q.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		if arg.Idforumtopic == 1 {
			return &db.GetForumTopicByIdForUserRow{Idforumtopic: 1, Title: sql.NullString{String: "Topic 1", Valid: true}}, nil
		}
		return nil, sql.ErrNoRows
	}
	q.SystemCheckGrantFn = func(p db.SystemCheckGrantParams) (int32, error) {
		return 1, nil
	}

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 1

	req := httptest.NewRequest("GET", "/topic/1/unread", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rec := httptest.NewRecorder()
	UnreadThreadsPage(rec, req)

	if val, ok := capturedTopicIDNull.(sql.NullInt32); ok {
		if !val.Valid || val.Int32 != 1 {
			t.Errorf("Expected TopicIDNull to be Valid and 1, got %v", val)
		}
	} else {
		t.Errorf("Expected TopicIDNull to be sql.NullInt32")
	}

	if capturedTopicIDVal != 1 {
		t.Errorf("Expected TopicIDVal to be 1, got %d", capturedTopicIDVal)
	}
}

func TestUnreadPrivateThreadsAllAccessible(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	var capturedTopicIDNull interface{}
	var capturedTopicIDVal int32

	q.ListUnreadPrivateThreadsForUserFn = func(ctx context.Context, arg db.ListUnreadPrivateThreadsForUserParams) ([]*db.ListUnreadPrivateThreadsForUserRow, error) {
		capturedTopicIDNull = arg.TopicIDNull
		capturedTopicIDVal = arg.TopicIDVal
		return nil, nil
	}

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 1

	req := httptest.NewRequest("GET", "/private/unread", nil)
	req = mux.SetURLVars(req, map[string]string{})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rec := httptest.NewRecorder()
	UnreadThreadsPage(rec, req)

	if val, ok := capturedTopicIDNull.(sql.NullInt32); ok {
		if val.Valid {
			t.Errorf("Expected TopicIDNull to be not Valid, got %v", val)
		}
	} else {
		if capturedTopicIDNull != nil {
			t.Errorf("Expected TopicIDNull to be nil or sql.NullInt32 with Valid=false, got %v", capturedTopicIDNull)
		}
	}

	if capturedTopicIDVal != 0 {
		t.Errorf("Expected TopicIDVal to be 0, got %d", capturedTopicIDVal)
	}
}
