package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"strconv"
)

func blogsBlogAddPage(w http.ResponseWriter, r *http.Request) {
	// TODO add guard
	type Data struct {
		*CoreData
		Languages          []*Language
		SelectedLanguageId int
		Mode               string
	}

	data := Data{
		CoreData:           r.Context().Value(ContextValues("coreData")).(*CoreData),
		SelectedLanguageId: 1,
		Mode:               "Add",
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	languageRows, err := queries.fetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	CustomBlogIndex(data.CoreData, r)

	if err := compiledTemplates.ExecuteTemplate(w, "blogsBlogAddPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func blogsBlogAddActionPage(w http.ResponseWriter, r *http.Request) {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	id, err := queries.add_blog(r.Context(), add_blogParams{
		UsersIdusers:       uid,
		LanguageIdlanguage: int32(languageId),
		Blog: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/blogs/blog/%d", id), http.StatusTemporaryRedirect)
}
