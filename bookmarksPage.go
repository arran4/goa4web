package goa4web

import (
	"log"
	"net/http"
)

func bookmarksPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	session, ok := GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	if uid == 0 {
		if err := renderTemplate(w, r, "bookmarksInfoPage.gohtml", data); err != nil {
			log.Printf("Template Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	bookmarksCustomIndex(data.CoreData)

	if err := renderTemplate(w, r, "bookmarksPage.gohtml", data); err != nil {
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
