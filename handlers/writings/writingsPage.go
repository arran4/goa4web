package writings

import (
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

func Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		CategoryId        int32
		WritingCategoryID int32
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Writings"
	data := Data{}
	data.CategoryId = 0
	data.WritingCategoryID = data.CategoryId

	handlers.TemplateHandler(w, r, "writingsPage", data)
}
func CustomWritingsIndex(data *common.CoreData, r *http.Request) {
	data.CustomIndexItems = []common.IndexItem{}

	data.CustomIndexItems = append(data.CustomIndexItems,
		common.IndexItem{Name: "Atom Feed", Link: "/writings/atom"},
		common.IndexItem{Name: "RSS Feed", Link: "/writings/rss"},
	)
	data.RSSFeedURL = "/writings/rss"
	data.AtomFeedURL = "/writings/atom"

	userHasAdmin := data.HasAdminRole() && data.AdminMode
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "User Roles",
			Link: "/admin/writings/users/roles",
		})
	}
	userHasWriter := data.HasContentWriterRole()
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
