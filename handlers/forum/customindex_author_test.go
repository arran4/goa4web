package forum

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
)

func TestCustomForumIndex_Author_NewStatus(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2/thread/3", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "thread": "3"})

	q := testhelpers.NewQuerierStub(
		testhelpers.WithGrant("forum", "topic", "reply"),
	)

	// Mock Thread details to simulate User 7 is the author
	q.GetThreadLastPosterAndPermsFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
		return &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:   3,
			Firstpostuserid: sql.NullInt32{Int32: 7, Valid: true},
		}, nil
	}

	// Mock Private labels: "unread" is inverted (read), "new" is missing (so implied new if author check fails)
	q.ListContentPrivateLabelsFn = func(arg db.ListContentPrivateLabelsParams) ([]*db.ListContentPrivateLabelsRow, error) {
		return []*db.ListContentPrivateLabelsRow{
			{Label: "unread", Invert: true},
			// "new" is missing
		}, nil
	}

	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))
	cd.UserID = 7

	CustomForumIndex(cd, req.WithContext(ctx))

	// If behavior is broken (passing 0 as author), "new" will be added because "new" label is missing in DB
	// and 0 != 7.
	// If "new" is added, hasThreadUnread returns true, so "Mark as read" link appears.

	// If behavior is fixed (passing 7 as author), "new" will NOT be added because 7 == 7.
	// "unread" is inverted, so it's not added.
	// hasThreadUnread returns false, so "Mark as read" link does NOT appear.

	hasMarkAsRead := common.ContainsItem(cd.CustomIndexItems, "Mark as read")
	if hasMarkAsRead {
		t.Error("Mark as read is present (Incorrect behavior)")
	}
}
