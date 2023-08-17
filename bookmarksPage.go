package main

import (
	"log"
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

	if err := getCompiledTemplates().ExecuteTemplate(w, "bookmarksPage.tmpl", data); err != nil {
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
