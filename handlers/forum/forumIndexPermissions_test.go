package forum

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/common/testdata"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
)

func TestCustomForumIndexWriteReply(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2/thread/3", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "thread": "3"})

	q := testhelpers.NewQuerierStub(
		testhelpers.WithGrant("forum", "topic", "reply"),
	)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))

	CustomForumIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "Write Reply") {
		t.Errorf("expected write reply item")
	}
	if len(q.SystemCheckGrantCalls) != 2 {
		t.Fatalf("expected 2 grant check, got %d", len(q.SystemCheckGrantCalls))
	}
}

func TestCustomForumIndexMarkReadLinks(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2/thread/3", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "thread": "3"})

	q := testhelpers.NewQuerierStub(
		testhelpers.WithGrant("forum", "topic", "reply"),
		testhelpers.WithPrivateLabels(testdata.VisibleThreadLabels(7)),
	)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))
	cd.UserID = 7

	CustomForumIndex(cd, req.WithContext(ctx))

	for _, name := range []string{"Mark as read", "Mark as read and go back", "Go to topic"} {
		if !common.ContainsItem(cd.CustomIndexItems, name) {
			t.Errorf("expected %s item", name)
		}
	}
	if len(q.ListContentPrivateLabelsCalls) != 1 {
		t.Fatalf("expected 1 private label query, got %d", len(q.ListContentPrivateLabelsCalls))
	}
}

func TestCustomForumIndexHidesMarkReadWhenClear(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2/thread/3", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "thread": "3"})

	q := testhelpers.NewQuerierStub(
		testhelpers.WithGrant("forum", "topic", "reply"),
		testhelpers.WithPrivateLabels([]*db.ListContentPrivateLabelsRow{
			{Item: "thread", ItemID: 3, UserID: 7, Label: "unread", Invert: true},
			{Item: "thread", ItemID: 3, UserID: 7, Label: "new", Invert: true},
		}),
	)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))
	cd.UserID = 7

	CustomForumIndex(cd, req.WithContext(ctx))

	for _, name := range []string{"Mark as read", "Mark as read and go back"} {
		if common.ContainsItem(cd.CustomIndexItems, name) {
			t.Errorf("unexpected %s item", name)
		}
	}
	if len(q.ListContentPrivateLabelsCalls) != 1 {
		t.Fatalf("expected 1 private label query, got %d", len(q.ListContentPrivateLabelsCalls))
	}
}

func TestCustomForumIndexWriteReplyDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2/thread/3", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "thread": "3"})

	q := testhelpers.NewQuerierStub(
		testhelpers.WithDefaultGrantAllowed(false),
	)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))

	CustomForumIndex(cd, req.WithContext(ctx))
	if common.ContainsItem(cd.CustomIndexItems, "Write Reply") {
		t.Errorf("unexpected write reply item")
	}
	if len(q.SystemCheckGrantCalls) != 2 {
		t.Fatalf("expected 2 grant check, got %d", len(q.SystemCheckGrantCalls))
	}
}

func TestCustomForumIndexCreateThread(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "category": "1"})

	q := testhelpers.NewQuerierStub(
		testhelpers.WithGrant("forum", "topic", "post"),
	)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))

	CustomForumIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "New Thread") {
		t.Errorf("expected create thread item")
	}
	if len(q.SystemCheckGrantCalls) != 2 {
		t.Fatalf("expected 2 grant check, got %d", len(q.SystemCheckGrantCalls))
	}
}

func TestCustomForumIndexAdminEditLink(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "category": "1"})

	q := testhelpers.NewQuerierStub(
		testhelpers.FromScenario(testhelpers.ScenarioAdmin()),
	)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	cd.UserID = 1
	cd.AdminMode = true

	CustomForumIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "Admin Edit Topic") {
		t.Errorf("expected admin edit link")
	}
}

func TestCustomForumIndexCreateThreadDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "category": "1"})

	q := testhelpers.NewQuerierStub(
		testhelpers.WithDefaultGrantAllowed(false),
	)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())

	CustomForumIndex(cd, req.WithContext(ctx))
	if common.ContainsItem(cd.CustomIndexItems, "New Thread") {
		t.Errorf("unexpected create thread item")
	}
	if len(q.SystemCheckGrantCalls) != 2 {
		t.Fatalf("expected 2 grant check, got %d", len(q.SystemCheckGrantCalls))
	}
}

func TestCustomForumIndexSubscribeLink(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "category": "1"})

	q := testhelpers.NewQuerierStub(
		testhelpers.WithGrant("forum", "topic", "see"),
	)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))
	cd.UserID = 1

	CustomForumIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "Subscribe To Topic") {
		t.Errorf("expected subscribe item")
	}
	if len(q.ListSubscriptionsByUserCalls) != 1 {
		t.Fatalf("expected 1 subscription query, got %d", len(q.ListSubscriptionsByUserCalls))
	}
}

func TestCustomForumIndexUnsubscribeLink(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "category": "1"})

	pattern := topicSubscriptionPattern(2)
	q := testhelpers.NewQuerierStub(
		testhelpers.WithGrant("forum", "topic", "see"),
		testhelpers.WithSubscriptions(testdata.SampleSubscriptions(1, pattern)),
	)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))
	cd.UserID = 1

	CustomForumIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "Unsubscribe From Topic") {
		t.Errorf("expected unsubscribe item")
	}
	if len(q.ListSubscriptionsByUserCalls) != 1 {
		t.Fatalf("expected 1 subscription query, got %d", len(q.ListSubscriptionsByUserCalls))
	}
}

func TestCustomForumIndexPrivateTopicAccess(t *testing.T) {
	q := testhelpers.NewQuerierStub(
		testhelpers.WithSubscriptions([]*db.ListSubscriptionsByUserRow{}),
	)
	q.SystemCheckGrantFn = func(p db.SystemCheckGrantParams) (int32, error) {
		if p.ViewerID == 1 && p.Section == "privateforum" && p.Item.String == "topic" && p.Action == "see" {
			if !p.ItemID.Valid || p.ItemID.Int32 == 0 {
				return 1, nil // Global grant
			}
		}
		// Allow post for topic 1 so the "New Thread" link appears (or create thread)
		if p.ViewerID == 1 && p.Section == "privateforum" && p.Item.String == "topic" && p.Action == "post" && p.ItemID.Int32 == 1 {
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

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))
	cd.UserID = 1

	// Test Topic 1 (Access Granted)
	req1 := httptest.NewRequest("GET", "/private/topic/1", nil)
	req1 = mux.SetURLVars(req1, map[string]string{"topic": "1", "category": "1"})

	// Add WithCoreData
	ctx1 := context.WithValue(req1.Context(), consts.KeyCoreData, cd)
	req1 = req1.WithContext(ctx1)

	CustomForumIndex(cd, req1)
	if !common.ContainsItem(cd.CustomIndexItems, "Subscribe To Topic") {
		t.Errorf("expected subscribe item for topic 1")
	}

	// Test Topic 2 (Access Denied)
	req2 := httptest.NewRequest("GET", "/private/topic/2", nil)
	req2 = mux.SetURLVars(req2, map[string]string{"topic": "2", "category": "1"})
	cd.CustomIndexItems = nil
	CustomForumIndex(cd, req2.WithContext(context.Background()))
	if common.ContainsItem(cd.CustomIndexItems, "Subscribe To Topic") {
		t.Errorf("unexpected subscribe item for topic 2")
	}
	if common.ContainsItem(cd.CustomIndexItems, "Unsubscribe From Topic") {
		t.Errorf("unexpected unsubscribe item for topic 2")
	}
	if common.ContainsItem(cd.CustomIndexItems, "Write Reply") {
		t.Errorf("unexpected write reply item for topic 2")
	}
}
