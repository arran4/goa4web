package main

import (
	"log"
	"net/http"
)

func blogsBlogPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	CustomIndex(data.CoreData, r)

	if err := compiledTemplates.ExecuteTemplate(w, "blogsBlogPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
