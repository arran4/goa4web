package main

import (
	"database/sql"
	"github.com/gorilla/sessions"
	"net/http"
	"strconv"
)

func faqAskPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Languages          []*Language
		SelectedLanguageId int32
	}

	data := Data{
		CoreData:           r.Context().Value(ContextValues("coreData")).(*CoreData),
		SelectedLanguageId: 1, // TODO user pref
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	CustomFAQIndex(data.CoreData)

	renderTemplate(w, r, "faqAskPage.gohtml", data)
}

func faqAskActionPage(w http.ResponseWriter, r *http.Request) {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	if err := queries.CreateFAQQuestion(r.Context(), CreateFAQQuestionParams{
		Question: sql.NullString{
			String: text,
			Valid:  true,
		},
		UsersIdusers:       uid,
		LanguageIdlanguage: int32(languageId),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	// TODO notify admin

	http.Redirect(w, r, "/faq", http.StatusTemporaryRedirect)
}
