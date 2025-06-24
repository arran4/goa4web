package goa4web

import (
	"database/sql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func writingsArticleAddPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Languages []*Language
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	CustomWritingsIndex(data.CoreData, r)

	if err := renderTemplate(w, r, "articleAddPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func writingsArticleAddActionPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryId, _ := strconv.Atoi(vars["category"])
	session, ok := GetSessionOrFail(w, r)
	if !ok {
		return
	}

	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	private, _ := strconv.ParseBool(r.PostFormValue("isitprivate"))
	title := r.PostFormValue("title")
	abstract := r.PostFormValue("abstract")
	body := r.PostFormValue("body")
	uid, _ := session.Values["UID"].(int32)

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	articleId, err := queries.InsertWriting(r.Context(), InsertWritingParams{
		WritingcategoryIdwritingcategory: int32(categoryId),
		Title:                            sql.NullString{Valid: true, String: title},
		Abstract:                         sql.NullString{Valid: true, String: abstract},
		Writting:                         sql.NullString{Valid: true, String: body},
		Private:                          sql.NullBool{Valid: true, Bool: private},
		LanguageIdlanguage:               int32(languageId),
		UsersIdusers:                     uid,
	})
	if err != nil {
		log.Printf("insertWriting Error: %s", err)
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

		if InsertWordsToWritingSearch(w, r, wordIds, queries, articleId) {
			return
		}
	}
}
