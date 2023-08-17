package main

import (
	"log"
	"net/http"
)

func taskDoneAutoRefreshPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	data.AutoRefresh = true

	if err := getCompiledTemplates().ExecuteTemplate(w, "taskDoneAutoRefreshPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
