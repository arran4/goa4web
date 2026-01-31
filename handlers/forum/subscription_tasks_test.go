package forum

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestSubscribeTopicTaskAction(t *testing.T) {
	qs := testhelpers.NewQuerierStub()
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(context.Background(), qs, cfg)
	cd.UserID = 42

	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test"
	sess := testhelpers.Must(store.Get(httptest.NewRequest(http.MethodGet, "http://example.com", nil), core.SessionName))
	sess.Values["UID"] = cd.UserID

	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)

	req := httptest.NewRequest(http.MethodPost, "http://example.com/forum/topic/12/subscribe", nil).WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"topic": "12"})
	rr := httptest.NewRecorder()

	result := subscribeTopicTaskAction.Action(rr, req)
	redirect, ok := result.(handlers.RedirectHandler)
	if !ok {
		t.Fatalf("expected RedirectHandler result, got %T", result)
	}
	if string(redirect) != "/forum/topic/12" {
		t.Fatalf("expected redirect to /forum/topic/12, got %q", string(redirect))
	}

	if len(qs.InsertSubscriptionParams) != 1 {
		t.Fatalf("expected 1 subscription insert, got %d", len(qs.InsertSubscriptionParams))
	}
	insert := qs.InsertSubscriptionParams[0]
	if insert.UsersIdusers != cd.UserID {
		t.Fatalf("expected user id %d, got %d", cd.UserID, insert.UsersIdusers)
	}
	if insert.Pattern != "create thread:/forum/topic/12/*" {
		t.Fatalf("expected subscription pattern %q, got %q", "create thread:/forum/topic/12/*", insert.Pattern)
	}
	if insert.Method != "internal" {
		t.Fatalf("expected subscription method internal, got %q", insert.Method)
	}
}

func TestUnsubscribeTopicTaskAction(t *testing.T) {
	qs := testhelpers.NewQuerierStub()
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(context.Background(), qs, cfg)
	cd.UserID = 24

	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test"
	sess := testhelpers.Must(store.Get(httptest.NewRequest(http.MethodGet, "http://example.com", nil), core.SessionName))
	sess.Values["UID"] = cd.UserID

	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)
	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)

	req := httptest.NewRequest(http.MethodPost, "http://example.com/forum/topic/9/unsubscribe", nil).WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"topic": "9"})
	rr := httptest.NewRecorder()

	result := unsubscribeTopicTaskAction.Action(rr, req)
	redirect, ok := result.(handlers.RedirectHandler)
	if !ok {
		t.Fatalf("expected RedirectHandler result, got %T", result)
	}
	if string(redirect) != "/forum/topic/9" {
		t.Fatalf("expected redirect to /forum/topic/9, got %q", string(redirect))
	}

	if len(qs.DeleteSubscriptionParams) != 1 {
		t.Fatalf("expected 1 subscription delete, got %d", len(qs.DeleteSubscriptionParams))
	}
	del := qs.DeleteSubscriptionParams[0]
	if del.SubscriberID != cd.UserID {
		t.Fatalf("expected subscriber id %d, got %d", cd.UserID, del.SubscriberID)
	}
	if del.Pattern != "create thread:/forum/topic/9/*" {
		t.Fatalf("expected subscription pattern %q, got %q", "create thread:/forum/topic/9/*", del.Pattern)
	}
	if del.Method != "internal" {
		t.Fatalf("expected subscription method internal, got %q", del.Method)
	}
}
