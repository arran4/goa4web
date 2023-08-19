package main

import (
	"log"
	"net/http"
)

func imagebbsBoardPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Boards []*printSubBoardsRow
		Posts  []*Post
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	CustomImageBBSIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "imagebbsBoardPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func imagebbsBoardPostImageActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO
	queries.addImage
	updateSearch
	taskDoneAutoRefreshPage
}
