package forum

import (
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

func TestCreateThreadLabels(t *testing.T) {
	creatorUID := int32(1)
	topicID := int32(5)
	newThreadID := int32(100)
	creationPagePath := fmt.Sprintf("/forum/topic/%d/thread", topicID)

	qs := testhelpers.NewQuerierStub()
	qs.GetPermissionsByUserIDFn = func(idusers int32) ([]*db.GetPermissionsByUserIDRow, error) {
		return []*db.GetPermissionsByUserIDRow{}, nil
	}
	qs.SystemGetUserByIDFn = func(ctx context.Context, idusers int32) (*db.SystemGetUserByIDRow, error) {
		return &db.SystemGetUserByIDRow{
			Idusers:  creatorUID,
			Username: sql.NullString{String: "creator", Valid: true},
			Email:    sql.NullString{String: "creator@example.com", Valid: true},
		}, nil
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
	// Mock for checking thread creation permission
	qs.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		return 1, nil
	}
	// Mock for HandleThreadUpdated -> notifications
	qs.GetThreadLastPosterAndPermsFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
		return &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:          newThreadID,
			ForumtopicIdforumtopic: topicID,
			Lastposterusername:     sql.NullString{String: "creator", Valid: true},
		}, nil
	}
	qs.GetPreferenceForListerReturn = map[int32]*db.Preference{
		creatorUID: {AutoSubscribeReplies: true},
	}
	qs.AdminListAdministratorEmailsReturns = []string{}

	bus := eventbus.NewBus()
	cfg := config.NewRuntimeConfig()
	cfg.NotificationsEnabled = false // Disable notifications for this test

	n := notifications.New(
		notifications.WithQueries(qs),
		notifications.WithConfig(cfg),
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
		Path:    creationPagePath,
		Task:    task,
		Outcome: eventbus.TaskOutcomeSuccess,
	}
	ctx := context.Background()
	cd := common.NewCoreData(ctx, qs, cfg, common.WithSession(sess), common.WithEvent(evt), common.WithUserRoles([]string{"member"}))
	cd.UserID = creatorUID

	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	form := url.Values{
		"replytext": {"First post content"},
		"language":  {"1"},
		"task":      {string(TaskCreateThread)},
		"public":    {"pub-label-1", "pub-label-2"},
		"private":   {"priv-label-1"},
	}
	req := httptest.NewRequest(http.MethodPost, "http://example.com"+creationPagePath, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"topic": fmt.Sprint(topicID)})

	rr := httptest.NewRecorder()

	cd.SetEvent(evt)
	cd.SetEventTask(task)

	task.Action(rr, req)

	// Verify labels were added
	if len(qs.AddContentPublicLabelCalls) != 2 {
		t.Errorf("Expected 2 public label calls, got %d", len(qs.AddContentPublicLabelCalls))
	} else {
		// The order isn't guaranteed by map iteration in SetPublicLabels if it iterates a map,
		// but SetThreadPublicLabels iterates the input slice if there are no existing labels.
		// Wait, SetPublicLabels fetches existing, calculates have/want, then adds/removes.
		// Since it's a new thread, there are no existing labels.
		// "want" is a map. So order is random.
		labels := make(map[string]bool)
		for _, call := range qs.AddContentPublicLabelCalls {
			if call.ItemID != newThreadID {
				t.Errorf("Expected ItemID %d, got %d", newThreadID, call.ItemID)
			}
			if call.Item != "thread" {
				t.Errorf("Expected Item 'thread', got %s", call.Item)
			}
			labels[call.Label] = true
		}
		if !labels["pub-label-1"] {
			t.Error("Missing pub-label-1")
		}
		if !labels["pub-label-2"] {
			t.Error("Missing pub-label-2")
		}
	}

	foundPrivateLabel := false
	for _, call := range qs.AddContentPrivateLabelCalls {
		if call.ItemID == newThreadID && call.Item == "thread" && call.Label == "priv-label-1" {
			foundPrivateLabel = true
			if call.UserID != creatorUID {
				t.Errorf("Expected UserID %d, got %d", creatorUID, call.UserID)
			}
		}
	}
	if !foundPrivateLabel {
		t.Error("Missing priv-label-1")
	}
}
