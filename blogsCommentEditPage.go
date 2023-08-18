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

func blogsCommentEditPostPage(w http.ResponseWriter, r *http.Request) {

	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("replytext")

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])
	commentId, _ := strconv.Atoi(vars["comment"])
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	comment, err := queries.getComment(r.Context(), int32(commentId))
	if err != nil {
		log.Printf("Error: getComment: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	thread, err := queries.user_get_thread(r.Context(), user_get_threadParams{
		UsersIdusers:  uid,
		Idforumthread: comment.ForumthreadIdforumthread,
	})
	if err != nil {
		log.Printf("Error: getComment: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err = queries.update_comment(r.Context(), update_commentParams{
		Idcomments:         int32(commentId),
		LanguageIdlanguage: int32(languageId),
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err := queries.update_forumthread(r.Context(), thread.Idforumthread); err != nil {
		log.Printf("Error: update_forumthread: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err := queries.update_forumtopic(r.Context(), thread.ForumtopicIdforumtopic); err != nil {
		log.Printf("Error: update_forumtopic: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/blogs/blog/%d/comments", blogId), http.StatusTemporaryRedirect)

}

func blogsCommentEditPostCancelPage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])
	http.Redirect(w, r, fmt.Sprintf("/blogs/blog/%d/comments", blogId), http.StatusTemporaryRedirect)

}
