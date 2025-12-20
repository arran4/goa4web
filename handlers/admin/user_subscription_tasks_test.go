package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

type subscriptionQueries struct {
	db.Querier
	inserted []db.InsertSubscriptionParams
	updated  []db.UpdateSubscriptionByIDForSubscriberParams
	deleted  []db.DeleteSubscriptionByIDForSubscriberParams
}

func (q *subscriptionQueries) InsertSubscription(_ context.Context, arg db.InsertSubscriptionParams) error {
	q.inserted = append(q.inserted, arg)
	return nil
}

func (q *subscriptionQueries) UpdateSubscriptionByIDForSubscriber(_ context.Context, arg db.UpdateSubscriptionByIDForSubscriberParams) error {
	q.updated = append(q.updated, arg)
	return nil
}

func (q *subscriptionQueries) DeleteSubscriptionByIDForSubscriber(_ context.Context, arg db.DeleteSubscriptionByIDForSubscriberParams) error {
	q.deleted = append(q.deleted, arg)
	return nil
}

func setupSubscriptionTaskTest(t *testing.T, userID int, body url.Values, queries db.Querier) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()
	var reader *strings.Reader
	if body != nil {
		reader = strings.NewReader(body.Encode())
	} else {
		reader = strings.NewReader("")
	}
	req := httptest.NewRequest("POST", "/admin/user/"+strconv.Itoa(userID)+"/subscriptions", reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req = mux.SetURLVars(req, map[string]string{"user": strconv.Itoa(userID)})
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), queries, cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	return httptest.NewRecorder(), req
}

func TestAddUserSubscriptionTask_UsesURLParam(t *testing.T) {
	body := url.Values{"pattern": {"/foo"}, "method": {"internal"}}
	queries := &subscriptionQueries{}
	rr, req := setupSubscriptionTaskTest(t, 9, body, queries)
	if err, ok := addUserSubscriptionTask.Action(rr, req).(error); ok && err != nil {
		t.Fatalf("Action: %v", err)
	}
	if len(queries.inserted) != 1 {
		t.Fatalf("expected insert, got %d", len(queries.inserted))
	}
	if arg := queries.inserted[0]; arg.UsersIdusers != 9 || arg.Pattern != "/foo" || arg.Method != "internal" {
		t.Fatalf("unexpected insert args: %#v", arg)
	}
}

func TestUpdateUserSubscriptionTask_UsesURLParam(t *testing.T) {
	body := url.Values{"id": {"3"}, "pattern": {"/bar"}, "method": {"email"}}
	queries := &subscriptionQueries{}
	rr, req := setupSubscriptionTaskTest(t, 4, body, queries)
	if err, ok := updateUserSubscriptionTask.Action(rr, req).(error); ok && err != nil {
		t.Fatalf("Action: %v", err)
	}
	if len(queries.updated) != 1 {
		t.Fatalf("expected update, got %d", len(queries.updated))
	}
	if arg := queries.updated[0]; arg.Pattern != "/bar" || arg.Method != "email" || arg.SubscriberID != 4 || arg.ID != 3 {
		t.Fatalf("unexpected update args: %#v", arg)
	}
}

func TestDeleteUserSubscriptionTask_UsesURLParam(t *testing.T) {
	body := url.Values{"id": {"5"}}
	queries := &subscriptionQueries{}
	rr, req := setupSubscriptionTaskTest(t, 11, body, queries)
	if err, ok := deleteUserSubscriptionTask.Action(rr, req).(error); ok && err != nil {
		t.Fatalf("Action: %v", err)
	}
	if len(queries.deleted) != 1 {
		t.Fatalf("expected delete, got %d", len(queries.deleted))
	}
	if arg := queries.deleted[0]; arg.SubscriberID != 11 || arg.ID != 5 {
		t.Fatalf("unexpected delete args: %#v", arg)
	}
}
