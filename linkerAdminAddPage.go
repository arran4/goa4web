package main

import (
	"database/sql"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"strconv"
)

func linkerAdminAddPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Languages  []*Language
		Categories []*Linkercategory
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	categoryRows, err := queries.showCategories(r.Context())
	if err != nil {
		log.Printf("forumCategories Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	data.Categories = categoryRows

	languageRows, err := queries.fetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	CustomLinkerIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "linkerAdminAddPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func linkerAdminAddActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)

	uid, _ := session.Values["UID"].(int32)

	title := r.PostFormValue("title")
	url := r.PostFormValue("URL")
	description := r.PostFormValue("description")
	category, _ := strconv.Atoi(r.PostFormValue("category"))

	if err := queries.addToLinker(r.Context(), addToLinkerParams{
		UsersIdusers:                   uid,
		LinkercategoryIdlinkercategory: int32(category),
		Title:                          sql.NullString{Valid: true, String: title},
		Url:                            sql.NullString{Valid: true, String: url},
		Description:                    sql.NullString{Valid: true, String: description},
	}); err != nil {
		log.Printf("addToLinker Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	taskDoneAutoRefreshPage(w, r)

}
