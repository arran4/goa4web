package main

import (
	"log"
	"net/http"
)

func imagebbsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	CustomImageBBSIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "imagebbsPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CustomImageBBSIndex(data *CoreData, r *http.Request) {

	data.RSSFeedUrl = "/imagebbs/rss"
	data.AtomFeedUrl = "/imagebbs/atom"

	userHasAdmin := true // TODO
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Admin",
			Link: "/admin",
		}, IndexItem{
			Name: "Modify Boards",
			Link: "/imagebbs/admin/boards",
		}, IndexItem{
			Name: "New Board",
			Link: "/imagebbs/admin/board",
		})
	}
}
