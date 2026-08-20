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

func TestPrivateForumCustomIndexPrivateTopicAccess(t *testing.T) {
	q := testhelpers.NewQuerierStub(
		testhelpers.WithSubscriptions([]*db.ListSubscriptionsByUserRow{}),
	)
	q.SystemCheckGrantFn = func(p db.SystemCheckGrantParams) (int32, error) {
		if p.ViewerID == 1 && p.Section == "privateforum" && p.Item.String == "topic" && p.Action == "see" {
			if !p.ItemID.Valid || p.ItemID.Int32 == 0 {
				return 1, nil
			}
		}
		if p.ViewerID == 1 && p.Section == "privateforum" && p.Item.String == "topic" && p.Action == "post" && p.ItemID.Int32 == 1 {
			return 1, nil
		}
		if p.ViewerID == 1 && p.Section == "privateforum" && p.Item.String == "topic" && p.Action == "edit" && p.ItemID.Int32 == 1 {
			return 1, nil
		}
		return 0, sql.ErrNoRows
	}
	q.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		if arg.ViewerID != 1 {
			return nil, sql.ErrNoRows
		}
		if arg.Idforumtopic == 1 {
			return &db.GetForumTopicByIdForUserRow{Idforumtopic: 1}, nil
		}
		return nil, sql.ErrNoRows
	}
	q.CountUnreadPrivateThreadsForUserFn = func(ctx context.Context, arg db.CountUnreadPrivateThreadsForUserParams) (int64, error) {
		if val, ok := arg.TopicIDNull.(sql.NullInt32); ok && val.Valid && val.Int32 == 1 {
			return 1, nil
		}
		if val, ok := arg.TopicIDNull.(sql.NullInt32); ok && val.Valid && val.Int32 == 2 {
			return 2, nil
		}
		return 0, nil
	}

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))
	cd.UserID = 1

	// Test Topic 1 (Access Granted)
	req1 := httptest.NewRequest("GET", "/private/topic/1/thread/1", nil)
	req1 = mux.SetURLVars(req1, map[string]string{"topic": "1", "thread": "1"})

	ctx1 := context.WithValue(req1.Context(), consts.KeyCoreData, cd)
	req1 = req1.WithContext(ctx1)

	CustomIndex(cd, req1)
	if !common.ContainsItem(cd.CustomIndexItems, "Edit Topic") {
		t.Errorf("expected Edit Topic item for topic 1")
	}
	if !common.ContainsItem(cd.CustomIndexItems, "Private threads") {
		t.Errorf("expected Private threads item for topic 1")
	}
	if !common.ContainsItem(cd.CustomIndexItems, "Unread in Topic (1)") {
		t.Errorf("expected Unread in Topic (1) item for topic 1")
	}
	if !common.ContainsItem(cd.CustomIndexItems, "Create a new private thread") {
		t.Errorf("expected Create a new private thread item for topic 1")
	}

	// Test Topic 2 (Access Denied)
	cd.CustomIndexItems = nil
	req2 := httptest.NewRequest("GET", "/private/topic/2/thread/1", nil)
	req2 = mux.SetURLVars(req2, map[string]string{"topic": "2", "thread": "1"})

	ctx2 := context.WithValue(req2.Context(), consts.KeyCoreData, cd)
	req2 = req2.WithContext(ctx2)

	CustomIndex(cd, req2)

	forbiddenItems := []string{
		"Edit Topic",
		"Private threads",
		"Unread in Topic",
		"Subscribe To Topic",
		"Unsubscribe From Topic",
		"New Thread",
		"Manage Labels",
		"Write Reply",
		"Create a new private thread",
	}

	for _, item := range forbiddenItems {
		if common.ContainsItem(cd.CustomIndexItems, item) {
			t.Errorf("unexpected %s item for topic 2", item)
		}
	}
}
