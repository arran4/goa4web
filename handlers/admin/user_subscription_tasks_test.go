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
	"github.com/arran4/goa4web/internal/testhelpers"
)

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

func TestHappyPathUserSubscriptionTasks(t *testing.T) {
	t.Run("Add Uses URL Param", func(t *testing.T) {
		body := url.Values{"pattern": {"/foo"}, "method": {"internal"}}
		q := testhelpers.NewQuerierStub()
		rr, req := setupSubscriptionTaskTest(t, 9, body, q)
		if err, ok := addUserSubscriptionTask.Action(rr, req).(error); ok && err != nil {
			t.Fatalf("Action: %v", err)
		}
		if len(q.InsertSubscriptionParams) != 1 {
			t.Fatalf("expected insert, got %d", len(q.InsertSubscriptionParams))
		}
		if arg := q.InsertSubscriptionParams[0]; arg.UsersIdusers != 9 || arg.Pattern != "/foo" || arg.Method != "internal" {
			t.Fatalf("unexpected insert args: %#v", arg)
		}
	})

	t.Run("Update Uses URL Param", func(t *testing.T) {
		body := url.Values{"id": {"3"}, "pattern": {"/bar"}, "method": {"email"}}
		q := testhelpers.NewQuerierStub()
		rr, req := setupSubscriptionTaskTest(t, 4, body, q)
		if err, ok := updateUserSubscriptionTask.Action(rr, req).(error); ok && err != nil {
			t.Fatalf("Action: %v", err)
		}
		if len(q.UpdateSubscriptionByIDForSubscriberCalls) != 1 {
			t.Fatalf("expected update, got %d", len(q.UpdateSubscriptionByIDForSubscriberCalls))
		}
		if arg := q.UpdateSubscriptionByIDForSubscriberCalls[0]; arg.Pattern != "/bar" || arg.Method != "email" || arg.SubscriberID != 4 || arg.ID != 3 {
			t.Fatalf("unexpected update args: %#v", arg)
		}
	})

	t.Run("Delete Uses URL Param", func(t *testing.T) {
		body := url.Values{"id": {"5"}}
		q := testhelpers.NewQuerierStub()
		rr, req := setupSubscriptionTaskTest(t, 11, body, q)
		if err, ok := deleteUserSubscriptionTask.Action(rr, req).(error); ok && err != nil {
			t.Fatalf("Action: %v", err)
		}
		if len(q.DeleteSubscriptionByIDForSubscriberCalls) != 1 {
			t.Fatalf("expected delete, got %d", len(q.DeleteSubscriptionByIDForSubscriberCalls))
		}
		if arg := q.DeleteSubscriptionByIDForSubscriberCalls[0]; arg.SubscriberID != 11 || arg.ID != 5 {
			t.Fatalf("unexpected delete args: %#v", arg)
		}
	})
}
