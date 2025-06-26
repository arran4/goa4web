package bookmarks

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core"
	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	common "github.com/arran4/goa4web/handlers/common"
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
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
	}

	if uid == 0 {
		if err := templates.RenderTemplate(w, "infoPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
			log.Printf("Template Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	bookmarksCustomIndex(data.CoreData)

	if err := templates.RenderTemplate(w, "bookmarksPage", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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
