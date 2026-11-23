package user

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

func TestNotificationsFeed(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/usr/notifications/rss", nil)
		cd := common.NewCoreData(context.Background(), nil, nil, common.WithAbsoluteURLBase("http://example.com"))
		ctx := context.WithValue(r.Context(), consts.KeyCoreData, cd)
		r = r.WithContext(ctx)

		n := []*db.Notification{{ID: 1, Link: sql.NullString{String: "/l", Valid: true}, Message: sql.NullString{String: "m", Valid: true}}}
		feed := NotificationsFeed(r, n, "Site")
		expectedLink := "http://example.com/usr/notifications/open/1"
		if len(feed.Items) != 1 || feed.Items[0].Link.Href != expectedLink {
			t.Fatalf("feed item incorrect, got %s want %s", feed.Items[0].Link.Href, expectedLink)
		}
		for _, write := range []struct {
			name string
			fn   func(io.Writer) error
		}{
			{name: "RSS", fn: feed.WriteRss},
			{name: "Atom", fn: feed.WriteAtom},
		} {
			t.Run(write.name, func(t *testing.T) {
				var output bytes.Buffer
				if err := write.fn(&output); err != nil {
					t.Fatalf("write feed: %v", err)
				}
				if !strings.Contains(output.String(), expectedLink) {
					t.Fatalf("serialized feed link does not contain %q", expectedLink)
				}
			})
		}

		router := mux.NewRouter()
		RegisterRoutes(router, nil)
		feedRequest := httptest.NewRequest(http.MethodGet, "/usr/notifications/open/1", nil)
		feedRequest = feedRequest.WithContext(context.WithValue(feedRequest.Context(), consts.KeyCoreData, common.NewCoreData(feedRequest.Context(), nil, nil).ForUser(1)))
		if !router.Match(feedRequest, &mux.RouteMatch{}) {
			t.Fatal("feed item URL is not registered for authenticated GET requests")
		}
		if feed.Title != "Site - Notifications" {
			t.Errorf("feed title incorrect: %s", feed.Title)
		}
	})
}
