package main

import (
	_ "embed"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver.
	"log"
	"net/http"
)

func adminUserPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	// AdminUserPermissionsData holds the data needed for rendering the template.
	type AdminUserPermissionsData struct {
		*CoreData
		Rows []*adminUserPermissionsRow
	}

	data := AdminUserPermissionsData{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	rows, err := queries.adminUserPermissions(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Rows = rows

	err = compiledTemplates.ExecuteTemplate(w, "adminUsersPermissionsPage.tmpl", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
