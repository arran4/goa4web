package forum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func TestThreadPageMarksAsRead(t *testing.T) {
	origStore := core.Store
	origName := core.SessionName
	core.Store = sessions.NewCookieStore([]byte("test"))
	core.SessionName = "test-session"
	defer func() {
		core.Store = origStore
		core.SessionName = origName
	}()

	threadID := int32(100)
	topicID := int32(200)
	lastCommentID := int32(300)

	queries := testhelpers.NewQuerierStub()

	// Mock required queries for page rendering
	queries.GetThreadLastPosterAndPermsFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
		return &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:          threadID,
			ForumtopicIdforumtopic: topicID,
			Firstpost:              1,
			Firstpostuserid:        sql.NullInt32{Int32: 1, Valid: true},
		}, nil
	}
	queries.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		return &db.GetForumTopicByIdForUserRow{
			Idforumtopic:                 topicID,
			Title:                        sql.NullString{String: "Topic Title", Valid: true},
			ForumcategoryIdforumcategory: 1,
		}, nil
	}
	queries.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		return []*db.GetCommentsByThreadIdForUserRow{
			{Idcomments: 1, Text: sql.NullString{String: "First post", Valid: true}, Written: sql.NullTime{Time: time.Now(), Valid: true}},
			{Idcomments: lastCommentID, Text: sql.NullString{String: "Last post", Valid: true}, Written: sql.NullTime{Time: time.Now(), Valid: true}},
		}, nil
	}

	// Track calls to mark read
	upsertReadMarkerCalled := false
	var capturedReadMarker db.UpsertContentReadMarkerParams
	queries.UpsertContentReadMarkerFn = func(ctx context.Context, arg db.UpsertContentReadMarkerParams) error {
		upsertReadMarkerCalled = true
		capturedReadMarker = arg
		return nil
	}

	addPrivateLabelCalled := 0
	capturedPrivateLabels := []db.AddContentPrivateLabelParams{}
	queries.AddContentPrivateLabelFn = func(ctx context.Context, arg db.AddContentPrivateLabelParams) error {
		addPrivateLabelCalled++
		capturedPrivateLabels = append(capturedPrivateLabels, arg)
		return nil
	}

	// Setup Request
	req := httptest.NewRequest("GET", "/forum/topic/200/thread/100", nil)
	req = mux.SetURLVars(req, map[string]string{
		"topic":  "200",
		"thread": "100",
	})

	// Setup Session
	sess, _ := core.Store.New(req, core.SessionName)
	sess.Values["UID"] = int32(1)
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), sess)

	// Setup CoreData
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Need to manually set current thread/topic as RequireThreadAndTopic middleware does
	cd.SetCurrentThreadAndTopic(threadID, topicID)

	// Invoke Handler
	ThreadPageWithBasePath(w, req, "/forum")

	// Assertions
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !upsertReadMarkerCalled {
		t.Error("Expected UpsertContentReadMarker to be called, but it wasn't")
	} else {
		if capturedReadMarker.ItemID != threadID {
			t.Errorf("Expected read marker thread ID %d, got %d", threadID, capturedReadMarker.ItemID)
		}
		if capturedReadMarker.LastCommentID != lastCommentID {
			t.Errorf("Expected read marker last comment ID %d, got %d", lastCommentID, capturedReadMarker.LastCommentID)
		}
	}

	// We expect AddContentPrivateLabel to be called twice (for "new" and "unread") with Invert: true
	foundUnread := false
	foundNew := false
	for _, call := range capturedPrivateLabels {
		if call.Item == "thread" && call.ItemID == threadID && call.Invert == true {
			if call.Label == "unread" {
				foundUnread = true
			}
			if call.Label == "new" {
				foundNew = true
			}
		}
	}

	if !foundUnread {
		t.Error("Expected AddContentPrivateLabel for 'unread' with Invert=true, but not found")
	}
	if !foundNew {
		t.Error("Expected AddContentPrivateLabel for 'new' with Invert=true, but not found")
	}
}
