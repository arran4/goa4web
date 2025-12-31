package forum

import (
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/common/testdata"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
)

func TestCustomForumIndexWriteReply(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2/thread/3", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "thread": "3"})

	q := testhelpers.NewQuerierStub(testhelpers.StubConfig{
		Grants: map[string]bool{
			testhelpers.GrantKey("forum", "topic", "reply"): true,
		},
	})
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

	q := testhelpers.NewQuerierStub(testhelpers.StubConfig{
		Grants: map[string]bool{
			testhelpers.GrantKey("forum", "topic", "reply"): true,
		},
		PrivateLabels: testdata.VisibleThreadLabels(7),
	})
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

	q := testhelpers.NewQuerierStub(testhelpers.StubConfig{
		Grants: map[string]bool{
			testhelpers.GrantKey("forum", "topic", "reply"): true,
		},
		PrivateLabels: []*db.ListContentPrivateLabelsRow{
			{Item: "thread", ItemID: 3, UserID: 7, Label: "unread", Invert: true},
			{Item: "thread", ItemID: 3, UserID: 7, Label: "new", Invert: true},
		},
	})
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

	q := testhelpers.NewQuerierStub(testhelpers.StubConfig{
		DefaultGrantAllowed: false,
	})
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

	q := testhelpers.NewQuerierStub(testhelpers.StubConfig{
		Grants: map[string]bool{
			testhelpers.GrantKey("forum", "topic", "post"): true,
		},
	})
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))

	CustomForumIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "New Thread") {
		t.Errorf("expected create thread item")
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected 1 grant check, got %d", len(q.SystemCheckGrantCalls))
	}
}

func TestCustomForumIndexCreateThreadFromThread(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2/thread/3", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "thread": "3"})

	q := testhelpers.NewQuerierStub(testhelpers.StubConfig{
		Grants: map[string]bool{
			testhelpers.GrantKey("forum", "topic", "post"): true,
		},
	})
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))

	CustomForumIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "New Thread") {
		t.Errorf("expected create thread item")
	}
	if len(q.SystemCheckGrantCalls) != 2 {
		t.Fatalf("Expected 2 grant checks, got %d", len(q.SystemCheckGrantCalls))
	}
}

func TestCustomForumIndexCreateThreadFromThreadPrivate(t *testing.T) {
	req := httptest.NewRequest("GET", "/private/topic/2/thread/3", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "thread": "3"})

	q := testhelpers.NewQuerierStub(testhelpers.StubConfig{
		Grants: map[string]bool{
			testhelpers.GrantKey("privateforum", "topic", "post"): true,
		},
	})
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))

	CustomForumIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "New Private Thread") {
		t.Errorf("expected create private thread item")
	}
}

func TestCustomForumIndexAdminEditLink(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "category": "1"})

	// Needed for IsAdminMode check
	q := testhelpers.NewQuerierStub(testhelpers.StubConfig{})

	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithPermissions([]*db.GetPermissionsByUserIDRow{
		{Name: "administrator", IsAdmin: true},
	}))
	cd.AdminMode = true

	CustomForumIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "Admin Edit Topic") {
		t.Errorf("expected admin edit link")
	}
}

func TestCustomForumIndexCreateThreadDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "category": "1"})

	q := testhelpers.NewQuerierStub(testhelpers.StubConfig{
		DefaultGrantAllowed: false,
	})
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())

	CustomForumIndex(cd, req.WithContext(ctx))
	if common.ContainsItem(cd.CustomIndexItems, "New Thread") {
		t.Errorf("unexpected create thread item")
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected 1 grant check, got %d", len(q.SystemCheckGrantCalls))
	}
}

func TestCustomForumIndexSubscribeLink(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "category": "1"})

	q := testhelpers.NewQuerierStub(testhelpers.StubConfig{})
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
	q := testhelpers.NewQuerierStub(testhelpers.StubConfig{
		Subscriptions: testdata.SampleSubscriptions(1, pattern),
	})
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
