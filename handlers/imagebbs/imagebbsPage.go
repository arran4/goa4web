package imagebbs

import (
	"log"
	"net/http"

	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

func Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*handlers.CoreData
		Boards      []*db.Imageboard
		IsSubBoard  bool
		BoardNumber int
	}

	data := Data{
		CoreData:    r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData),
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

func CustomImageBBSIndex(data *handlers.CoreData, r *http.Request) {

	if data.FeedsEnabled {
		data.RSSFeedUrl = "/imagebbs/rss"
		data.AtomFeedUrl = "/imagebbs/atom"
	}

	userHasAdmin := data.HasRole("administrator") && data.AdminMode
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, corecommon.IndexItem{
			Name: "Admin",
			Link: "/admin",
		}, corecommon.IndexItem{
			Name: "Modify Boards",
			Link: "/admin/imagebbs/boards",
		}, corecommon.IndexItem{
			Name: "New Board",
			Link: "/admin/imagebbs/board",
		})
	}
}
