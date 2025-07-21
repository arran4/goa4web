package bookmarks

import (
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

func Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
	}

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}

	if uid == 0 {
		handlers.TemplateHandler(w, r, "infoPage.gohtml", data)
		return
	}

	handlers.TemplateHandler(w, r, "bookmarksPage", data)
}

func bookmarksCustomIndex(data *common.CoreData, r *http.Request) {
	data.CustomIndexItems = []common.IndexItem{}
	data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
		Name: "Show",
		Link: "/bookmarks/mine",
	})
	data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
		Name: "Edit",
		Link: "/bookmarks/edit",
	})
}
