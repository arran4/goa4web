package main

import (
	"log"
	"net/http"
)

func adminForumFlaggedPostsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct{ *CoreData }
	data := Data{CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData)}
	if err := renderTemplate(w, r, "adminForumFlaggedPostsPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminForumModeratorLogsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct{ *CoreData }
	data := Data{CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData)}
	if err := renderTemplate(w, r, "adminForumModeratorLogsPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
