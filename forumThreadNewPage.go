package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"strconv"
)

func forumThreadNewPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Languages          []*Language
		SelectedLanguageId int
	}

	data := Data{
		CoreData:           r.Context().Value(ContextValues("coreData")).(*CoreData),
		SelectedLanguageId: 1, // TODO update these from user prefs and make it an optional filter
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	languageRows, err := queries.fetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	CustomBlogIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "forumThreadNewPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func forumThreadNewActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	topicId, err := strconv.Atoi(vars["topic"])
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	// TODO check if the user has the right right to topic

	threadId, err := queries.makeThread(r.Context(), int32(topicId))
	if err != nil {
		log.Printf("Error: makeThread: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d", topicId, threadId)

	if rows, err := queries.threadNotify(r.Context(), threadNotifyParams{
		ForumthreadIdforumthread: int32(threadId),
		Idusers:                  uid,
	}); err != nil {
		log.Printf("Error: threadNotify: %s", err)
	} else {
		for _, row := range rows {
			if err := notifyChange(r.Context(), getEmailProvider(), row.String, endUrl); err != nil {
				log.Printf("Error: notifyChange: %s", err)
			}
		}
	}

	if rows, err := queries.threadNotify(r.Context(), threadNotifyParams{
		Idusers:                  uid,
		ForumthreadIdforumthread: int32(threadId),
	}); err != nil {
		log.Printf("Error: threadNotify: %s", err)
	} else {
		for _, row := range rows {
			if err := notifyChange(r.Context(), getEmailProvider(), row.String, endUrl); err != nil {
				log.Printf("Error: notifyChange: %s", err)

			}
		}
	}

	if err := queries.makePost(r.Context(), makePostParams{
		LanguageIdlanguage:       int32(languageId),
		UsersIdusers:             uid,
		ForumthreadIdforumthread: int32(threadId),
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	}); err != nil {
		log.Printf("Error: makeThread: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err := queries.update_forumthread(r.Context(), int32(threadId)); err != nil {
		log.Printf("Error: update_forumthread: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err := queries.update_forumtopic(r.Context(), int32(topicId)); err != nil {
		log.Printf("Error: update_forumtopic: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}

func forumThreadNewCancelPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])

	endUrl := fmt.Sprintf("/forum/topic/%d", topicId)

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}
