package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
)

func linkerSuggestPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Categories         []*Linkercategory
		Languages          []*Language
		SelectedLanguageId int
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	categoryRows, err := queries.GetAllLinkerCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllForumCategories Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	data.Categories = categoryRows

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	CustomLinkerIndex(data.CoreData, r)
	if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "linkerSuggestPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func linkerSuggestActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	session, ok := GetSessionOrFail(w, r)
	if !ok {
		return
	}

	uid, _ := session.Values["UID"].(int32)

	title := r.PostFormValue("title")
	url := r.PostFormValue("URL")
	description := r.PostFormValue("description")
	category, _ := strconv.Atoi(r.PostFormValue("category"))

	if err := queries.CreateLinkerQueuedItem(r.Context(), CreateLinkerQueuedItemParams{
		UsersIdusers:                   uid,
		LinkercategoryIdlinkercategory: int32(category),
		Title:                          sql.NullString{Valid: true, String: title},
		Url:                            sql.NullString{Valid: true, String: url},
		Description:                    sql.NullString{Valid: true, String: description},
	}); err != nil {
		log.Printf("createLinkerQueuedItem Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	taskDoneAutoRefreshPage(w, r)
}
