package forum

import (
	"github.com/arran4/goa4web/handlers/forum/forumcommon"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestCreateThreadNotificationLink(t *testing.T) {
	creatorUID := int32(1)
	adminUID := int32(99)
	topicID := int32(5)
	newThreadID := int32(100)
	createdThreadPath := fmt.Sprintf("/forum/topic/%d/thread/%d", topicID, newThreadID)
	creationPagePath := fmt.Sprintf("/forum/topic/%d/thread", topicID)

	qs := testhelpers.NewQuerierStub()
	qs.GetPermissionsByUserIDFn = func(idusers int32) ([]*db.GetPermissionsByUserIDRow, error) {
		return []*db.GetPermissionsByUserIDRow{}, nil
	}
	qs.SystemGetUserByIDFn = func(ctx context.Context, idusers int32) (*db.SystemGetUserByIDRow, error) {
		switch idusers {
		case creatorUID:
			return &db.SystemGetUserByIDRow{
				Idusers:  creatorUID,
				Username: sql.NullString{String: "creator", Valid: true},
				Email:    sql.NullString{String: "creator@example.com", Valid: true},
			}, nil
		case adminUID:
			return &db.SystemGetUserByIDRow{
				Idusers:  adminUID,
				Username: sql.NullString{String: "adminuser", Valid: true},
				Email:    sql.NullString{String: "admin@example.com", Valid: true},
			}, nil
		}
		return nil, sql.ErrNoRows
	}
	qs.SystemGetUserByEmailFn = func(ctx context.Context, email string) (*db.SystemGetUserByEmailRow, error) {
		if email == "admin@example.com" {
			return &db.SystemGetUserByEmailRow{Idusers: adminUID}, nil
		}
		return nil, sql.ErrNoRows
	}
	qs.SystemCreateThreadFn = func(ctx context.Context, idforumtopic int32) (int64, error) {
		return int64(newThreadID), nil
	}
	qs.CreateCommentInSectionForCommenterFn = func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
		return 999, nil
	}
	qs.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		return &db.GetForumTopicByIdForUserRow{
			Idforumtopic: topicID,
			Title:        sql.NullString{String: "Test Topic", Valid: true},
		}, nil
	}
	qs.GetPreferenceForListerReturn = map[int32]*db.Preference{
		creatorUID: {AutoSubscribeReplies: true},
	}
	qs.AdminListAdministratorEmailsReturns = []string{"admin@example.com"}

	// Mocks for HasGrant/UserCanCreateThread checks
	qs.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		// Allow everything
		return 1, nil
	}
	qs.GetThreadLastPosterAndPermsFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
		return &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:          newThreadID,
			ForumtopicIdforumtopic: topicID,
			Lastposterusername:     sql.NullString{String: "creator", Valid: true},
		}, nil
	}
	// Mock for CoreData.Languages() which is called in Page but maybe not Action?
	// Action calls r.PostFormValue("language")

	// CoreData.Languages() is not called in Action.

	bus := eventbus.NewBus()
	cfg := config.NewRuntimeConfig()
	cfg.NotificationsEnabled = true
	cfg.AdminNotify = true

	mockProvider := &MockEmailProvider{}
	n := notifications.New(
		notifications.WithQueries(qs),
		notifications.WithConfig(cfg),
		notifications.WithEmailProvider(mockProvider),
	)
	cdlq := &captureDLQ{}
	n.RegisterSync(bus, cdlq)

	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test"
	sess, _ := store.Get(httptest.NewRequest(http.MethodGet, "http://example.com", nil), core.SessionName)
	sess.Values["UID"] = creatorUID

	task := CreateThreadTaskHandler
	evt := &eventbus.TaskEvent{
		Data:    map[string]any{},
		UserID:  creatorUID,
		Path:    creationPagePath, // The path of the request
		Task:    task,
		Outcome: eventbus.TaskOutcomeSuccess,
	}
	ctx := context.Background()
	// Need to initialize CoreData with event
	cd := common.NewCoreData(ctx, qs, cfg, common.WithSession(sess), common.WithEvent(evt), common.WithUserRoles([]string{"member"}))
	cd.UserID = creatorUID

	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	form := url.Values{
		"replytext": {"First post content"},
		"language":  {"1"},
		"task":      {string(forumcommon.TaskCreateThread)},
	}
	req := httptest.NewRequest(http.MethodPost, "http://example.com"+creationPagePath, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"topic": fmt.Sprint(topicID)})

	rr := httptest.NewRecorder()

	// Execute the action (wrapped in TaskHandler logic partially here by setting up context/CoreData)
	// But we call Action directly. Note that TaskHandler usually sets cd.Event.Task.
	// We set it manually in evt.

	// Ensure CoreData has event
	cd.SetEvent(evt)
	cd.SetEventTask(task)

	task.Action(rr, req)

	// Trigger synchronous processing of the event
	bus.Publish(*evt)

	if cdlq.lastError != "" {
		t.Errorf("sync process error: %s", cdlq.lastError)
	}

	// Check the notifications created
	foundAdminNotification := false
	for _, call := range qs.SystemCreateNotificationCalls {
		if call.RecipientID == adminUID {
			foundAdminNotification = true
			if call.Link.String != createdThreadPath {
				t.Errorf("Expected admin notification link to be %q, got %q", createdThreadPath, call.Link.String)
			}
		}
	}
	if !foundAdminNotification {
		t.Fatal("Did not find admin notification")
	}
}
