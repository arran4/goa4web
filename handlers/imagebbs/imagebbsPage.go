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
		data.RSSFeedURL = "/imagebbs/rss"
		data.AtomFeedURL = "/imagebbs/atom"
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
