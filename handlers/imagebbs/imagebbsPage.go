package imagebbs

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Boards      []*db.Imageboard
		IsSubBoard  bool
		BoardNumber int
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	handlers.SetPageTitle(r, "Image Board")
	data := Data{
		CoreData:    cd,
		IsSubBoard:  false,
		BoardNumber: 0,
	}

	boards, err := data.CoreData.SubImageBoards(0)
	if err != nil {
		log.Printf("imageboards: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.Boards = boards

	handlers.TemplateHandler(w, r, "imagebbsPage", data)
}

func CustomImageBBSIndex(data *common.CoreData, r *http.Request) {
	data.CustomIndexItems = []common.IndexItem{}

	if data.FeedsEnabled {
		data.RSSFeedUrl = "/imagebbs/rss"
		data.AtomFeedUrl = "/imagebbs/atom"
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
			Link: "/admin/imagebbs/board",
		})
	}
}
