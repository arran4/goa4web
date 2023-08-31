package main

import (
	"log"
	"net/http"
)

func imagebbsAdminPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	CustomImageBBSIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "imagebbsAdminPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
