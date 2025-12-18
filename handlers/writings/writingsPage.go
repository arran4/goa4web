package writings

import (
	"github.com/arran4/goa4web/core/common"
	"net/http"
)

func WritingsPage(w http.ResponseWriter, r *http.Request) {
	t := NewWritingsTask().(*writingsTask)
	t.Get(w, r)
}
func CustomWritingsIndex(data *common.CoreData, r *http.Request) {
	data.CustomIndexItems = []common.IndexItem{}

	data.CustomIndexItems = append(data.CustomIndexItems,
		common.IndexItem{Name: "Atom Feed", Link: "/writings/atom", Folded: true},
		common.IndexItem{Name: "RSS Feed", Link: "/writings/rss", Folded: true},
	)
	data.RSSFeedURL = "/writings/rss"
	data.AtomFeedURL = "/writings/atom"

	if data.IsAdmin() {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Writings Admin",
			Link: "/admin/writings",
		})
	}
	userHasWriter := data.HasContentWriterRole()
	if userHasWriter {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Write writings",
			Link: "/writings/add",
		})
	}

	data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
		Name: "Writers",
		Link: "/writings/writers",
	})

	data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
		Name: "Return to list",
		Link: "/writings?offset=0",
	})
}
