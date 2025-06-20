package main

import (
	"database/sql"
	"errors"
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

	comment, err := queries.GetCommentById(r.Context(), int32(commentId))
	if err != nil {
		log.Printf("Error: getComment: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	thread, err := queries.GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissions(r.Context(), GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsParams{
		UsersIdusers:  uid,
		Idforumthread: comment.ForumthreadIdforumthread,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("Error: getThreadByIdForUserByIdWithLastPoserUserNameAndPermissions: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	if err = queries.UpdateComment(r.Context(), UpdateCommentParams{
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

	if err := PostUpdate(r.Context(), queries, thread.Idforumthread, thread.ForumtopicIdforumtopic); err != nil {
		log.Printf("Error: postUpdate: %s", err)
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
