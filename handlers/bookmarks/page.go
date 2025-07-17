package bookmarks

import (
	"net/http"

	"github.com/arran4/goa4web/core"
	corecommon "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
)

func Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
	}

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	data := Data{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*corecommon.CoreData),
	}

	if uid == 0 {
		handlers.TemplateHandler(w, r, "infoPage.gohtml", data)
		return
	}

	handlers.TemplateHandler(w, r, "bookmarksPage", data)
}

func bookmarksCustomIndex(data *corecommon.CoreData) {
	data.CustomIndexItems = append(data.CustomIndexItems, corecommon.IndexItem{
		Name: "Show",
		Link: "/bookmarks/mine",
	})
	data.CustomIndexItems = append(data.CustomIndexItems, corecommon.IndexItem{
		Name: "Edit",
		Link: "/bookmarks/edit",
	})
}
