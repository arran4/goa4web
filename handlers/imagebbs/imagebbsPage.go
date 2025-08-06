package imagebbs

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

func Page(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Image Board"
	handlers.TemplateHandler(w, r, "imagebbsPage", struct{}{})
}

func CustomImageBBSIndex(data *common.CoreData, r *http.Request) {
	data.CustomIndexItems = []common.IndexItem{}

	if data.FeedsEnabled {
		data.RSSFeedURL = "/imagebbs/rss"
		data.AtomFeedURL = "/imagebbs/atom"
	}

	userHasAdmin := data.HasRole("administrator") && data.AdminMode
	if userHasAdmin {
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
