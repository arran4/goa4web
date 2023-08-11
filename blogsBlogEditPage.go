package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func blogsBlogEditPage(w http.ResponseWriter, r *http.Request) {
	// TODO add guard
	type Data struct {
		*CoreData
		Languages          []*Language
		Blog               *show_blog_editRow
		SelectedLanguageId int
		Mode               string
	}

	data := Data{
		CoreData:           r.Context().Value(ContextValues("coreData")).(*CoreData),
		SelectedLanguageId: 1,
		Mode:               "Edit",
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	languageRows, err := queries.fetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])

	row, err := queries.show_blog_edit(r.Context(), int32(blogId))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Blog = row

	CustomBlogIndex(data.CoreData, r)

	if err := compiledTemplates.ExecuteTemplate(w, "blogsBlogEditPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func blogsBlogEditActionPage(w http.ResponseWriter, r *http.Request) {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])

	err = queries.update_blog(r.Context(), update_blogParams{
		Idblogs:            int32(blogId),
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

	http.Redirect(w, r, fmt.Sprintf("/blogs/blog/%d", blogId), http.StatusTemporaryRedirect)
}
