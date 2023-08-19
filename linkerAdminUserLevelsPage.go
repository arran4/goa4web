package main

import (
	"log"
	"net/http"
)

func linkerAdminUserLevelsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	// Custom Index???

	CustomLinkerIndex(data.CoreData, r)
	if err := getCompiledTemplates().ExecuteTemplate(w, "linkerAdminUserLevelsPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func linkerAdminUserLevelsAllowActionPage(w http.ResponseWriter, r *http.Request) {
}

func linkerAdminUserLevelsRemoveActionPage(w http.ResponseWriter, r *http.Request) {
}
