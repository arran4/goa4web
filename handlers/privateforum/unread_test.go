package privateforum

import (
	"context"
	"database/sql"
	"net/http"
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
	var listCalls int
	var capturedTopicIDNull interface{}
	var capturedTopicIDVal int32

	q.ListUnreadPrivateThreadsForUserFn = func(ctx context.Context, arg db.ListUnreadPrivateThreadsForUserParams) ([]*db.ListUnreadPrivateThreadsForUserRow, error) {
		listCalls++
		capturedTopicIDNull = arg.TopicIDNull
		capturedTopicIDVal = arg.TopicIDVal
		return []*db.ListUnreadPrivateThreadsForUserRow{
			{
				Idforumthread: 101,
				TopicID:       1,
				TopicTitle:    sql.NullString{String: "Private Topic 1", Valid: true},
			},
		}, nil
	}
	q.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		if arg.Idforumtopic == 1 {
			return &db.GetForumTopicByIdForUserRow{
				Idforumtopic: 1,
				Title:        sql.NullString{String: "Private Topic 1", Valid: true},
			}, nil
		}
		return nil, sql.ErrNoRows
	}
	q.SystemCheckGrantFn = func(p db.SystemCheckGrantParams) (int32, error) {
		if p.ViewerID == 1 && p.Section == "privateforum" && p.Item.String == "topic" && p.Action == "see" {
			return 1, nil
		}
		return 0, sql.ErrNoRows
	}

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 1

	req := httptest.NewRequest("GET", "/private/topic/1/unread", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rec := httptest.NewRecorder()
	UnreadThreadsPage(rec, req)

	if listCalls != 1 {
		t.Fatalf("Expected ListUnreadPrivateThreadsForUser to be called exactly once, got %d", listCalls)
	}

	if val, ok := capturedTopicIDNull.(sql.NullInt32); ok {
		if !val.Valid || val.Int32 != 1 {
			t.Errorf("Expected TopicIDNull to be Valid and 1, got %v", val)
		}
	} else {
		t.Errorf("Expected TopicIDNull to be sql.NullInt32, got %T", capturedTopicIDNull)
	}

	if capturedTopicIDVal != 1 {
		t.Errorf("Expected TopicIDVal to be 1, got %d", capturedTopicIDVal)
	}
}

func TestUnreadPrivateThreadsAllAccessible(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	var listCalls int
	var capturedTopicIDNull interface{}
	var capturedTopicIDVal int32

	q.SystemCheckGrantFn = func(p db.SystemCheckGrantParams) (int32, error) {
		if p.ViewerID == 1 && p.Section == "privateforum" && p.Item.String == "topic" && p.Action == "see" {
			return 1, nil
		}
		return 0, sql.ErrNoRows
	}

	q.ListUnreadPrivateThreadsForUserFn = func(ctx context.Context, arg db.ListUnreadPrivateThreadsForUserParams) ([]*db.ListUnreadPrivateThreadsForUserRow, error) {
		listCalls++
		capturedTopicIDNull = arg.TopicIDNull
		capturedTopicIDVal = arg.TopicIDVal
		return []*db.ListUnreadPrivateThreadsForUserRow{
			{
				Idforumthread: 101,
				TopicID:       1,
				TopicTitle:    sql.NullString{String: "Topic 1", Valid: true},
			},
			{
				Idforumthread: 102,
				TopicID:       3,
				TopicTitle:    sql.NullString{String: "Topic 3", Valid: true},
			},
		}, nil
	}

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 1

	req := httptest.NewRequest("GET", "/private/unread", nil)
	req = mux.SetURLVars(req, map[string]string{})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rec := httptest.NewRecorder()
	UnreadThreadsPage(rec, req)

	if listCalls != 1 {
		t.Fatalf("Expected ListUnreadPrivateThreadsForUser to be called exactly once, got %d", listCalls)
	}

	if val, ok := capturedTopicIDNull.(sql.NullInt32); ok {
		if val.Valid {
			t.Errorf("Expected TopicIDNull to be Valid=false for unscoped unread, got %v", val)
		}
	} else if capturedTopicIDNull != nil {
		t.Errorf("Expected TopicIDNull to be sql.NullInt32 with Valid=false or nil, got %v", capturedTopicIDNull)
	}

	if capturedTopicIDVal != 0 {
		t.Errorf("Expected TopicIDVal to be 0 for unscoped unread, got %d", capturedTopicIDVal)
	}
}

func TestUnreadPrivateThreadsInaccessibleTopicDenied(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	var listCalls int

	q.SystemCheckGrantFn = func(p db.SystemCheckGrantParams) (int32, error) {
		if p.ViewerID == 1 && p.Section == "privateforum" && p.Item.String == "topic" && p.Action == "see" {
			return 1, nil
		}
		return 0, sql.ErrNoRows
	}

	q.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		// Topic 2 is inaccessible (no view access)
		return nil, sql.ErrNoRows
	}

	q.ListUnreadPrivateThreadsForUserFn = func(ctx context.Context, arg db.ListUnreadPrivateThreadsForUserParams) ([]*db.ListUnreadPrivateThreadsForUserRow, error) {
		listCalls++
		return nil, nil
	}

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 1

	req := httptest.NewRequest("GET", "/private/topic/2/unread", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2"})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rec := httptest.NewRecorder()
	UnreadThreadsPage(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404 Not Found for inaccessible topic unread, got %d", rec.Code)
	}

	if listCalls != 0 {
		t.Errorf("Expected 0 calls to ListUnreadPrivateThreadsForUser on access denial, got %d", listCalls)
	}
}
