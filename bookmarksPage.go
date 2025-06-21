package main

import (
	"net/http"
)

func bookmarksPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	bookmarksCustomIndex(data.CoreData)

	renderTemplate(w, r, "bookmarksPage.gohtml", data)
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
