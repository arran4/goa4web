package forum

import (
	"context"
	"database/sql"
	"fmt"
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
	"github.com/gorilla/mux"
)

func TestManageTopicLabelsPage(t *testing.T) {
	t.Run("renders correctly for authorized user", func(t *testing.T) {
		topicID := int32(1)
		queries := testhelpers.NewQuerierStub()

		// Mock Permission Check
		queries.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
			if arg.Section == "forum" && arg.Action == "label" && arg.ItemID.Int32 == topicID {
				return 1, nil
			}
			return 0, sql.ErrNoRows
		}

		// Mock Topic Fetch
		queries.GetForumTopicByIdFn = func(ctx context.Context, id int32) (*db.Forumtopic, error) {
			if id == topicID {
				return &db.Forumtopic{
					Idforumtopic: topicID,
					Title:        sql.NullString{String: "Test Topic", Valid: true},
					Lastaddition: sql.NullTime{Time: time.Now(), Valid: true},
				}, nil
			}
			return nil, sql.ErrNoRows
		}

		// cd.ForumTopicByID tries GetForumTopicByIdForUser if GetForumTopicById fails or ... actually it probably tries GetForumTopicById if admin, or GetForumTopicByIdForUser if user?
		// But let's mock GetForumTopicByIdForUser as well just in case.
		queries.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
			if arg.Idforumtopic == topicID {
				return &db.GetForumTopicByIdForUserRow{
					Idforumtopic: topicID,
					Title:        sql.NullString{String: "Test Topic", Valid: true},
					Lastaddition: sql.NullTime{Time: time.Now(), Valid: true},
				}, nil
			}
			return nil, sql.ErrNoRows
		}

		// Mock Labels Fetch
		queries.ListContentPublicLabelsFn = func(arg db.ListContentPublicLabelsParams) ([]*db.ListContentPublicLabelsRow, error) {
			return []*db.ListContentPublicLabelsRow{
				{Label: "Label1", Item: "topic", ItemID: topicID},
				{Label: "Label2", Item: "topic", ItemID: topicID},
			}, nil
		}
		queries.ListContentLabelStatusReturns = []*db.ListContentLabelStatusRow{}
		queries.ListContentPrivateLabelsFn = func(arg db.ListContentPrivateLabelsParams) ([]*db.ListContentPrivateLabelsRow, error) {
			return []*db.ListContentPrivateLabelsRow{
				{Label: "PrivateLabel1", Item: "topic", ItemID: topicID, UserID: 123},
			}, nil
		}

		// Setup CoreData
		cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
		cd.UserID = 123

		// Setup Request
		req := httptest.NewRequest("GET", fmt.Sprintf("/forum/topic/%d/labels", topicID), nil)
		req = mux.SetURLVars(req, map[string]string{"topic": fmt.Sprintf("%d", topicID)})
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		ManageTopicLabelsPage(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Manage Labels for Topic: Test Topic") {
			t.Errorf("expected title in body, got: %s", body)
		}
		if !strings.Contains(body, "Label1") {
			t.Errorf("expected Label1 in body")
		}
		if !strings.Contains(body, "Label2") {
			t.Errorf("expected Label2 in body")
		}
		if !strings.Contains(body, "PrivateLabel1") {
			t.Errorf("expected PrivateLabel1 in body")
		}
	})

	t.Run("denies access for unauthorized user", func(t *testing.T) {
		topicID := int32(1)
		queries := testhelpers.NewQuerierStub()

		// Mock Permission Check to fail
		queries.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
			return 0, sql.ErrNoRows
		}

		cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
		cd.UserID = 123

		req := httptest.NewRequest("GET", fmt.Sprintf("/forum/topic/%d/labels", topicID), nil)
		req = mux.SetURLVars(req, map[string]string{"topic": fmt.Sprintf("%d", topicID)})
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		ManageTopicLabelsPage(w, req)

		if w.Code == http.StatusOK {
			t.Fatalf("expected error status, got 200")
		}
	})
}
