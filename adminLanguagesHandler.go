package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver.
	"net/http"
	"strconv"
)

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

	renderTemplate(w, r, "adminLanguagesPage.gohtml", data)
}

func adminLanguagesRenamePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	cid := r.PostFormValue("cid")
	cname := r.PostFormValue("cname")
	data := struct {
		*CoreData
		Errors []string
		Back   string
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
	renderTemplate(w, r, "adminRunTaskPage.gohtml", data)
}
func adminLanguagesDeletePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	cid := r.PostFormValue("cid")
	data := struct {
		*CoreData
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/languages",
	}
	if cidi, err := strconv.Atoi(cid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.DeleteLanguage(r.Context(), int32(cidi)); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteLanguage: %w", err).Error())
	}
	renderTemplate(w, r, "adminRunTaskPage.gohtml", data)
}
func adminLanguagesCreatePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	cname := r.PostFormValue("cname")
	data := struct {
		*CoreData
		Errors []string
		Back   string
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
	renderTemplate(w, r, "adminRunTaskPage.gohtml", data)
}
