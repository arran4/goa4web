package goa4web

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/common"
)

// adminPermissionsSectionViewPage lists all permissions for a specific section.
func adminPermissionsSectionViewPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Section string
		Rows    []*PermissionWithUser
	}
	cd := r.Context().Value(ContextValues("coreData")).(*CoreData)
	section := r.URL.Query().Get("section")
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	rows, err := queries.GetPermissionsBySectionWithUsers(r.Context(), section)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("GetPermissionsBySectionWithUsers error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := Data{CoreData: cd, Section: section, Rows: rows}
	if err := templates.RenderTemplate(w, "permissionsSectionViewPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
