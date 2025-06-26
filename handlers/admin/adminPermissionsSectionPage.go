package admin

import (
	"database/sql"
	"errors"
	"fmt"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

// AdminPermissionsSectionPage displays the current distinct permission sections
// in the database so administrators can verify whether "writing" or "writings"
// is in use.
func AdminPermissionsSectionPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Sections []*db.CountPermissionSectionsRow
	}
	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	rows, err := queries.CountPermissionSections(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("CountPermissionSections error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Sections = rows

	if err := templates.RenderTemplate(w, "permissionsSectionPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// AdminPermissionsSectionRenamePage converts one permission section value to
// another. This can be used to normalise "writing" vs "writings" values.
func AdminPermissionsSectionRenamePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	from := r.PostFormValue("from")
	to := r.PostFormValue("to")
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
		Back:     "/admin/permissions/sections",
	}

	if from == "" || to == "" {
		data.Errors = append(data.Errors, "from and to values required")
	} else if err := queries.RenamePermissionSection(r.Context(), db.RenamePermissionSectionParams{
		Section:   sql.NullString{String: to, Valid: true},
		Section_2: sql.NullString{String: from, Valid: true},
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RenamePermissionSection: %w", err).Error())
	}

	if err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
