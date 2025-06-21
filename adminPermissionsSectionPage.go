package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
)

// adminPermissionsSectionPage displays the current distinct permission sections
// in the database so administrators can verify whether "writing" or "writings"
// is in use.
func adminPermissionsSectionPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Sections []*CountPermissionSectionsRow
	}
	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	rows, err := queries.CountPermissionSections(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("CountPermissionSections error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Sections = rows

	if err := renderTemplate(w, r, "adminPermissionsSectionPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// adminPermissionsSectionRenamePage converts one permission section value to
// another. This can be used to normalise "writing" vs "writings" values.
func adminPermissionsSectionRenamePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	from := r.PostFormValue("from")
	to := r.PostFormValue("to")
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/permissions/sections",
	}

	if from == "" || to == "" {
		data.Errors = append(data.Errors, "from and to values required")
	} else if err := queries.RenamePermissionSection(r.Context(), RenamePermissionSectionParams{
		Section:   sql.NullString{String: to, Valid: true},
		Section_2: sql.NullString{String: from, Valid: true},
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RenamePermissionSection: %w", err).Error())
	}

	if err := renderTemplate(w, r, "adminRunTaskPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
