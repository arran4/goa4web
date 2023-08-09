package main

import (
	"log"
	"net/http"
)

func blogsBloggersPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	CustomIndex(data.CoreData, r)

	if err := compiledTemplates.ExecuteTemplate(w, "blogsBloggersPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
