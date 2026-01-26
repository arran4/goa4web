package user

import (
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

func TestNotificationsFeed(t *testing.T) {
	r := httptest.NewRequest("GET", "/usr/notifications/rss", nil)
	n := []*db.Notification{{ID: 1, Link: sql.NullString{String: "/l", Valid: true}, Message: sql.NullString{String: "m", Valid: true}}}
	feed := NotificationsFeed(r, n, "Site")
	if len(feed.Items) != 1 || feed.Items[0].Link.Href != "/l" {
		t.Fatalf("feed item incorrect")
	}
	if feed.Title != "Site - Notifications" {
		t.Errorf("feed title incorrect: %s", feed.Title)
	}
}
