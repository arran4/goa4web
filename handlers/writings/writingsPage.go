package writings

import (
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

var writingsPermissionsPageEnabled = true

func Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Categories        []*db.WritingCategory
		EditingCategoryId int32
		CategoryId        int32
		WritingCategoryID int32
		IsAdmin           bool
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}

	data.IsAdmin = data.CoreData.HasRole("administrator") && data.CoreData.AdminMode
	editID, _ := strconv.Atoi(r.URL.Query().Get("edit"))
	data.EditingCategoryId = int32(editID)
	data.CategoryId = 0
	data.WritingCategoryID = data.CategoryId

	categoryRows, err := data.CoreData.VisibleWritingCategories(data.CoreData.UserID)
	if err != nil {
		log.Printf("writingCategories: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Categories = append(data.Categories, categoryRows...)

	handlers.TemplateHandler(w, r, "writingsPage", data)
}
func CustomWritingsIndex(data *common.CoreData, r *http.Request) {
	data.CustomIndexItems = []common.IndexItem{}

	data.CustomIndexItems = append(data.CustomIndexItems,
		common.IndexItem{Name: "Atom Feed", Link: "/writings/atom"},
		common.IndexItem{Name: "RSS Feed", Link: "/writings/rss"},
	)
	data.RSSFeedUrl = "/writings/rss"
	data.AtomFeedUrl = "/writings/atom"

	userHasAdmin := data.HasRole("administrator") && data.AdminMode
	if userHasAdmin && writingsPermissionsPageEnabled {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "User Permissions",
			Link: "/writings/user/permissions",
		})
	}
	userHasWriter := data.HasRole("content writer")
	if userHasWriter || userHasAdmin {
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
		Link: fmt.Sprintf("/writings?offset=%d", 0),
	})
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset != 0 {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "The start",
			Link: fmt.Sprintf("/writings?offset=%d", 0),
		})
	}
	data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
		Name: "Next 10",
		Link: fmt.Sprintf("/writings?offset=%d", offset+10),
	})
	if offset > 0 {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Previous 10",
			Link: fmt.Sprintf("/writings?offset=%d", offset-10),
		})
	}
}
