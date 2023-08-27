package main

import (
	"log"
	"net/http"
)

func linkerCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Categories []*Linkercategory
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	categories, err := queries.showCategories(r.Context())
	if err != nil {
		log.Printf("showCategories Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.Categories = categories

	CustomLinkerIndex(data.CoreData, r)
	if err := getCompiledTemplates().ExecuteTemplate(w, "linkerCategoriesPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
