package main

import (
	"log"
	"net/http"
)

func linkerAdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	// Custom Index???
	CustomLinkerIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "linkerAdminCategoriesPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func linkerAdminCategoriesUpdatePage(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func linkerAdminCategoriesRenamePage(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func linkerAdminCategoriesDeletePage(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func linkerAdminCategoriesCreatePage(w http.ResponseWriter, r *http.Request) {
	// TODO
}
