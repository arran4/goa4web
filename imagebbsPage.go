package main

import (
	"log"
	"net/http"
)

func imagebbsHandler(w http.ResponseWriter, r *http.Request) {
	// Data holds the data needed for rendering the template.
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	if err := compiledTemplates.ExecuteTemplate(w, "imagebbsPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
