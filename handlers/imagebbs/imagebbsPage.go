package imagebbs

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
)

func ImagebbsPage(w http.ResponseWriter, r *http.Request) {
	t := NewImagebbsTask().(*imagebbsTask)
	t.Get(w, r)
}

func CustomImageBBSIndex(data *common.CoreData, r *http.Request) {
	data.CustomIndexItems = []common.IndexItem{}

	if data.FeedsEnabled {
		data.RSSFeedURL = data.GenerateFeedURL("/imagebbs/rss")
		data.RSSFeedTitle = "ImageBBS RSS Feed"
		data.AtomFeedURL = data.GenerateFeedURL("/imagebbs/atom")
		data.AtomFeedTitle = "ImageBBS Atom Feed"
		data.CustomIndexItems = append(data.CustomIndexItems,
			common.IndexItem{Name: "ImageBBS Atom Feed", Link: data.AtomFeedURL, Folded: true},
			common.IndexItem{Name: "ImageBBS RSS Feed", Link: data.RSSFeedURL, Folded: true},
		)
	}

	if data.IsAdmin() {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Admin",
			Link: "/admin",
		}, common.IndexItem{
			Name: "Modify Boards",
			Link: "/admin/imagebbs/boards",
		}, common.IndexItem{
			Name: "New Board",
			Link: "/admin/imagebbs/boards/new",
		})
	}
}
