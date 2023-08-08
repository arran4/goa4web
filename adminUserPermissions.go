package main

import (
	_ "embed"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver.
	"log"
	"net/http"
)

// Define your indexitem struct.
type indexitem struct {
	Name string // Name of URL displayed in <a href>
	Link string // URL for link.
}

// AdminUserPermissionsData holds the data needed for rendering the template.
type AdminUserPermissionsData struct {
	*CoreData
	Rows []*adminUserPermissionsRow
}

func adminUserPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(ContextValues("coreData")).(*CoreData)
	if cd.SecurityLevel != "all" {
		http.Error(w, "Incorrect security level", http.StatusForbidden)
		return
	}
	// Prepare the index items.
	data := AdminUserPermissionsData{
		CoreData: cd,
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	rows, err := queries.adminUserPermissions(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Rows = rows

	err = compiledTemplates.ExecuteTemplate(w, "adminUserPermissionsPage.tmpl", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
