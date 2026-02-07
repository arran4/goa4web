package user

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestNotificationsFeed(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/usr/notifications/rss", nil)
		cd := common.NewCoreData(context.Background(), nil, nil, common.WithAbsoluteURLBase("http://example.com"))
		ctx := context.WithValue(r.Context(), consts.KeyCoreData, cd)
		r = r.WithContext(ctx)

		n := []*db.Notification{{ID: 1, Link: sql.NullString{String: "/l", Valid: true}, Message: sql.NullString{String: "m", Valid: true}}}
		feed := NotificationsFeed(r, n, "Site")
		expectedLink := "http://example.com/usr/notifications/go/1"
		if len(feed.Items) != 1 || feed.Items[0].Link.Href != expectedLink {
			t.Fatalf("feed item incorrect, got %s want %s", feed.Items[0].Link.Href, expectedLink)
		}
		if feed.Title != "Site - Notifications" {
			t.Errorf("feed title incorrect: %s", feed.Title)
		}
	})
}
