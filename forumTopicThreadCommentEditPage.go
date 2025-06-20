package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func forumTopicThreadCommentEditActionPage(w http.ResponseWriter, r *http.Request) {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("replytext")

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])
	threadId, _ := strconv.Atoi(vars["thread"])
	commentId, _ := strconv.Atoi(vars["comment"])

	err = queries.UpdateComment(r.Context(), UpdateCommentParams{
		Idcomments:         int32(commentId),
		LanguageIdlanguage: int32(languageId),
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err := PostUpdate(r.Context(), queries, int32(threadId), int32(topicId)); err != nil {
		log.Printf("Error: postUpdate: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/forum/topic/%d/thread/%d#comment-%d", topicId, threadId, commentId), http.StatusTemporaryRedirect)
}

func forumTopicThreadCommentEditActionCancelPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])
	threadId, _ := strconv.Atoi(vars["thread"])

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicId, threadId)

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}
