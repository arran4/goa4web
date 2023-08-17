package main

import (
	_ "embed"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver.
	"log"
	"net/http"
)

func adminPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	err := getCompiledTemplates().ExecuteTemplate(w, "adminPage.tmpl", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
