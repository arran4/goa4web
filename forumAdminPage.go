package main

import (
	"log"
	"net/http"
)

func forumAdminPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	CustomForumIndex(data.CoreData, r)

	if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "forumAdminPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
