package goa4web

import (
	"github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
)

func bookmarksPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	if uid == 0 {
		if err := templates.RenderTemplate(w, "infoPage.gohtml", data, common.NewFuncs(r)); err != nil {
			log.Printf("Template Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	bookmarksCustomIndex(data.CoreData)

	if err := templates.RenderTemplate(w, "page.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func bookmarksCustomIndex(data *CoreData) {
	data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
		Name: "Show",
		Link: "/bookmarks/mine",
	})
	data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
		Name: "Edit",
		Link: "/bookmarks/edit",
	})
}
