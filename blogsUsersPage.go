package main

import (
	"log"
	"net/http"
)

func blogsUsersPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	CustomIndex(data.CoreData, r)

	if err := compiledTemplates.ExecuteTemplate(w, "blogsUsersPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
