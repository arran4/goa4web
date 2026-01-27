package user

import (
	"net/http"
	"strconv"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/feeds"
)

// NotificationsFeed converts a list of notifications into a feed.
func NotificationsFeed(r *http.Request, notifications []*db.Notification) *feeds.Feed {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	feed := &feeds.Feed{
		Title:       "Notifications",
		Link:        &feeds.Link{Href: r.URL.Path},
		Description: "recent notifications",
		Created:     time.Now(),
	}
	for _, n := range notifications {
		link := ""
		if n.Link.Valid {
			link = cd.AbsoluteURL("/usr/notifications/go/" + strconv.Itoa(int(n.ID)))
		}
		msg := ""
		if n.Message.Valid {
			msg = n.Message.String
		}
		item := &feeds.Item{
			Title:       msg,
			Link:        &feeds.Link{Href: link},
			Created:     time.Now(),
			Description: msg,
		}
		item.Created = n.CreatedAt
		feed.Items = append(feed.Items, item)
	}
	return feed
}
