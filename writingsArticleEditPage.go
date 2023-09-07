package main

import (
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"strconv"
)

func writingsArticleEditPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Languages          []*Language
		SelectedLanguageId int
		Writing            *GetWritingByIdForUserDescendingByPublishedDateRow
		UserId             int32
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	vars := mux.Vars(r)
	articleId, _ := strconv.Atoi(vars["article"])

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)
	data.UserId = uid

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	writing, err := queries.GetWritingByIdForUserDescendingByPublishedDate(r.Context(), GetWritingByIdForUserDescendingByPublishedDateParams{
		Userid:    uid,
		Idwriting: int32(articleId),
	})
	if err != nil {
		log.Printf("getWritingByIdForUserDescendingByPublishedDate Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.Writing = writing

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	CustomWritingsIndex(data.CoreData, r)

	if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "writingsArticleEditPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func writingsArticleEditActionPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	articleId, _ := strconv.Atoi(vars["article"])

	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	private, _ := strconv.ParseBool(r.PostFormValue("isitprivate"))
	title := r.PostFormValue("title")
	abstract := r.PostFormValue("abstract")
	body := r.PostFormValue("body")

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	if err := queries.UpdateWriting(r.Context(), UpdateWritingParams{
		Title:              sql.NullString{Valid: true, String: title},
		Abstract:           sql.NullString{Valid: true, String: abstract},
		Writting:           sql.NullString{Valid: true, String: body},
		Private:            sql.NullBool{Valid: true, Bool: private},
		LanguageIdlanguage: int32(languageId),
		Idwriting:          int32(articleId),
	}); err != nil {
		log.Printf("updateWriting Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := queries.WritingSearchDelete(r.Context(), int32(articleId)); err != nil {
		log.Printf("writingSearchDelete Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for _, text := range []string{
		abstract,
		title,
		body,
	} {
		wordIds, done := SearchWordIdsFromText(w, r, text, queries)
		if done {
			return
		}

		if InsertWordsToWritingSearch(w, r, wordIds, queries, int64(articleId)) {
			return
		}
	}
}
