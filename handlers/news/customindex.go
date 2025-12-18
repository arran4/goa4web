package news

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/gorilla/mux"
)

// NewsCustomIndexItems returns the context-aware index items for news pages.
func NewsCustomIndexItems(cd *common.CoreData, r *http.Request) []common.IndexItem {
	items := []common.IndexItem{}

	path := "/news"
	cd.RSSFeedURL = cd.GenerateFeedURL(path + ".rss")
	cd.PublicRSSFeedURL = path + ".rss"
	items = append(items, common.IndexItem{
		Name: "RSS Feed",
		Link: cd.RSSFeedURL,
	})

	if cd.HasGrant("news", "post", "post", 0) {
		items = append(items, common.IndexItem{
			Name: "Add News",
			Link: "/news/post",
		})
	}

	vars := mux.Vars(r)
	newsIDStr := vars["news"]

	if newsIDStr != "" {
		items = append(items, common.IndexItem{
			Name: "Return to list",
			Link: fmt.Sprintf("/?offset=%d", 0),
		})

		newsPost := cd.CurrentNewsPostLoaded()
		if newsPost != nil {
			newsID, _ := strconv.Atoi(newsIDStr)
			// Mark as read
			items = append(items,
				common.IndexItem{
					Name: "Mark as read",
					Link: markNewsReadLink(int32(newsID), r.URL.RequestURI()),
				},
				common.IndexItem{
					Name: "Mark as read and go back",
					Link: markNewsReadLink(int32(newsID), "/news"),
				},
			)
		}
	}

	return items
}

func markNewsReadLink(newsID int32, redirect string) string {
	link := fmt.Sprintf("/news/news/%d/labels?task=%s", newsID, url.QueryEscape("Mark Thread Read"))
	if redirect != "" {
		link = fmt.Sprintf("%s&redirect=%s", link, url.QueryEscape(redirect))
	}
	return link
}
