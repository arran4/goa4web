package admin

import (
	"database/sql"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

// AdminPermissionsSectionViewPage lists all permissions for a specific section.
func AdminPermissionsSectionViewPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Section string
		Rows    []*PermissionWithUser
	}
	cd := r.Context().Value(common.KeyCoreData).(*CoreData)
	section := r.URL.Query().Get("section")
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	rows, err := queries.GetPermissionsBySectionWithUsers(r.Context(), section)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("GetPermissionsBySectionWithUsers error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := Data{CoreData: cd, Section: section, Rows: rows}
	if err := templates.RenderTemplate(w, "permissionsSectionViewPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
