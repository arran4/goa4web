package forum

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

func TestForumCustomIndexItems_UnreadStatus(t *testing.T) {
	// Scenario: User is the author of the thread.
	// Current behavior (Bug): "Mark as read" appears because authorID is passed as 0, so it looks like someone else's thread.
	// Desired behavior (Fix): "Mark as read" should NOT appear because authorID matches UserID.

	userID := int32(10)
	threadID := int32(100)
	topicID := int32(1)

	queries := testhelpers.NewQuerierStub()

	// Mock getting thread details to return the author
	queries.GetThreadLastPosterAndPermsFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
		return &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:          threadID,
			Firstpostuserid:        sql.NullInt32{Int32: userID, Valid: true}, // User is the author
			ForumtopicIdforumtopic: topicID,
		}, nil
	}

	// Mock Private Labels.
	// The stub function signature for ListContentPrivateLabelsFn does not include context.Context
	queries.ListContentPrivateLabelsFn = func(arg db.ListContentPrivateLabelsParams) ([]*db.ListContentPrivateLabelsRow, error) {
		// Simulate that the user has read the thread (unread is inverted)
		// If we don't do this, "unread" label is added by default, making hasThreadUnread return true regardless of "new" label.
		return []*db.ListContentPrivateLabelsRow{
			{
				Label:  "unread",
				Invert: true,
			},
		}, nil
	}

    // We also need grants
	// The stub function signature for SystemCheckGrantFn does not include context.Context
    queries.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
        return 1, nil // Has grant
    }

    // Also need GetForumTopicByIdForUser for the link generation in ForumCustomIndexItems -> cd.RSSFeedURL part
    // Actually ForumCustomIndexItems calls cd.FeedsEnabled which is false by default.
    // It calls cd.IsAdmin(), subscribedToTopic(cd, int32(tid)) ...

    // subscribedToTopic calls cd.ListSubscriptionsByUser
    queries.ListSubscriptionsByUserReturns = []*db.ListSubscriptionsByUserRow{}

	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	cd.UserID = userID
    cd.FeedsEnabled = false

	req := httptest.NewRequest("GET", "/forum/topic/1/thread/100", nil)
	req = mux.SetURLVars(req, map[string]string{
		"topic":  "1",
		"thread": "100",
	})
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	items := ForumCustomIndexItems(cd, req)

	hasMarkAsRead := false
	for _, item := range items {
		if item.Name == "Mark as read" {
			hasMarkAsRead = true
			break
		}
	}

	if hasMarkAsRead {
		t.Error("'Mark as read' option should not be present for the author")
	}
}
