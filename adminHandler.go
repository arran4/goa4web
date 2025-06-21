package main

import (
	_ "embed"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver.
	"net/http"
)

func adminPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	renderTemplate(w, r, "adminPage.gohtml", data)
}
