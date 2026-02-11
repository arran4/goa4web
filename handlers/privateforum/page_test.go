package privateforum

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestUnhappyPathPage_NoAccess(t *testing.T) {
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
	cd.ShareSignKey = "secret"
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	w := httptest.NewRecorder()
	PrivateForumPage(w, req)

	if body := w.Body.String(); !strings.Contains(body, "Login Required") {
		t.Fatalf("expected login required message, got %q", body)
	}
}

func TestHappyPathPage_Access(t *testing.T) {
	queries := testhelpers.NewQuerierStub(testhelpers.WithGrantResult(true))
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig(), common.WithPrivateForumTopics(nil))
	cd.ShareSignKey = "secret"
	cd.UserID = 1
	cd.AdminMode = true
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	w := httptest.NewRecorder()
	PrivateForumPage(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "Private Topics") {
		t.Fatalf("expected private topics page, got %q", body)
	}
	if !strings.Contains(body, "topic-controls") {
		t.Fatalf("expected create form, got %q", body)
	}
}

func TestHappyPathPage_SeeNoCreate(t *testing.T) {
	q := testhelpers.NewQuerierStub(
		testhelpers.WithDefaultGrantAllowed(true),
		testhelpers.WithGrants(map[string]bool{
			testhelpers.GrantKey("privateforum", "", "post"): false,
		}),
	)
	q.GetPermissionsByUserIDReturns = []*db.GetPermissionsByUserIDRow{}
	q.ListContentPublicLabelsCalls = []db.ListContentPublicLabelsParams{}
	q.ListContentPrivateLabelsCalls = []db.ListContentPrivateLabelsParams{}
	q.ListPrivateTopicParticipantsByTopicIDForUserReturns = []*db.ListPrivateTopicParticipantsByTopicIDForUserRow{}
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig(), common.WithPrivateForumTopics(nil))
	cd.ShareSignKey = "secret"
	cd.UserID = 1
	cd.AdminMode = false

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	w := httptest.NewRecorder()
	PrivateForumPage(w, req)

	body := w.Body.String()
	if strings.Contains(body, "Start conversation") {
		t.Fatalf("unexpected create form, got %q", body)
	}
}

func TestHappyPathPage_AdminLinks(t *testing.T) {
	queries := testhelpers.NewQuerierStub(testhelpers.WithGrantResult(true))
	queries.GetPermissionsByUserIDReturns = []*db.GetPermissionsByUserIDRow{{IsAdmin: true}}

	// Inject a mock topic
	topic := &common.PrivateTopic{
		ListPrivateTopicsByUserIDRow: &db.ListPrivateTopicsByUserIDRow{},
	}
	topic.Idforumtopic = 123
	topic.Title.String = "Secret Plans"
	topic.Title.Valid = true
	topic.Lastaddition.Time = time.Now()
	topic.Lastaddition.Valid = true

	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig(), common.WithPrivateForumTopics([]*common.PrivateTopic{topic}))
	cd.ShareSignKey = "secret"
	cd.UserID = 1
	cd.AdminMode = true

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	w := httptest.NewRecorder()
	PrivateForumPage(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "/admin/forum/topics/topic/123/edit") {
		t.Errorf("expected admin link for topic 123, got body length %d", len(body))
	}
}
