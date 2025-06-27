package imagebbs

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

func Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Boards      []*db.Imageboard
		IsSubBoard  bool
		BoardNumber int
	}

	data := Data{
		CoreData:    r.Context().Value(common.KeyCoreData).(*common.CoreData),
		IsSubBoard:  false,
		BoardNumber: 0,
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	subBoardRows, err := queries.GetAllBoardsByParentBoardId(r.Context(), 0)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllBoardsByParentBoardId Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Boards = subBoardRows

	CustomImageBBSIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "imagebbsPage", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CustomImageBBSIndex(data *common.CoreData, r *http.Request) {

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
