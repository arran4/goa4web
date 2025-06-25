package goa4web

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func forumTopicThreadCommentEditActionPage(w http.ResponseWriter, r *http.Request) {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("replytext")

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	threadRow := r.Context().Value(ContextValues("thread")).(*GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsRow)
	topicRow := r.Context().Value(ContextValues("topic")).(*GetForumTopicByIdForUserRow)
	commentId, _ := strconv.Atoi(mux.Vars(r)["comment"])

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

	if err := PostUpdate(r.Context(), queries, threadRow.Idforumthread, topicRow.Idforumtopic); err != nil {
		log.Printf("Error: postUpdate: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/forum/topic/%d/thread/%d#comment-%d", topicRow.Idforumtopic, threadRow.Idforumthread, commentId), http.StatusTemporaryRedirect)
}

func forumTopicThreadCommentEditActionCancelPage(w http.ResponseWriter, r *http.Request) {
	threadRow := r.Context().Value(ContextValues("thread")).(*GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsRow)
	topicRow := r.Context().Value(ContextValues("topic")).(*GetForumTopicByIdForUserRow)

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicRow.Idforumtopic, threadRow.Idforumthread)

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}
