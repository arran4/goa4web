package goa4web

import (
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver.
)

func adminLanguageRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin/languages", http.StatusMovedPermanently)
}

func adminLanguagesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Rows []*Language
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	rows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Rows = rows

	err = templates.RenderTemplate(w, "languagesPage.gohtml", data, common.NewFuncs(r))
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminLanguagesRenamePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	cid := r.PostFormValue("cid")
	cname := r.PostFormValue("cname")
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/languages",
	}
	if cidi, err := strconv.Atoi(cid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.RenameLanguage(r.Context(), RenameLanguageParams{
		Nameof:     sql.NullString{Valid: true, String: cname},
		Idlanguage: int32(cidi),
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RenameLanguage: %w", err).Error())
	}
	err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, common.NewFuncs(r))
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func adminLanguagesDeletePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	cid := r.PostFormValue("cid")
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/languages",
	}
	if cidi, err := strconv.Atoi(cid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.DeleteLanguage(r.Context(), int32(cidi)); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteLanguage: %w", err).Error())
	}
	err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, common.NewFuncs(r))
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func adminLanguagesCreatePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	cname := r.PostFormValue("cname")
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/languages",
	}
	if err := queries.CreateLanguage(r.Context(), sql.NullString{
		String: cname,
		Valid:  true,
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("CreateLanguage: %w", err).Error())
	}
	err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, common.NewFuncs(r))
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
