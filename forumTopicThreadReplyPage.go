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

func forumTopicThreadReplyPage(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)

	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])
	threadId, _ := strconv.Atoi(vars["thread"])

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	uid, _ := session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicId, threadId)

	if rows, err := queries.ListUsersSubscribedToThread(r.Context(), ListUsersSubscribedToThreadParams{
		ForumthreadIdforumthread: int32(threadId),
		Idusers:                  uid,
	}); err != nil {
		log.Printf("Error: listUsersSubscribedToThread: %s", err)
	} else {
		for _, row := range rows {
			if err := notifyChange(r.Context(), getEmailProvider(), row.String, endUrl); err != nil {
				log.Printf("Error: notifyChange: %s", err)
			}
		}
	}

	if rows, err := queries.ListUsersSubscribedToThread(r.Context(), ListUsersSubscribedToThreadParams{
		Idusers:                  uid,
		ForumthreadIdforumthread: int32(threadId),
	}); err != nil {
		log.Printf("Error: listUsersSubscribedToThread: %s", err)
	} else {
		for _, row := range rows {
			if err := notifyChange(r.Context(), getEmailProvider(), row.String, endUrl); err != nil {
				log.Printf("Error: notifyChange: %s", err)

			}
		}
	}

	cid, err := queries.CreateComment(r.Context(), CreateCommentParams{
		LanguageIdlanguage:       int32(languageId),
		UsersIdusers:             uid,
		ForumthreadIdforumthread: int32(threadId),
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		log.Printf("Error: makeThread: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	/* TODO
	-- name: postUpdate :exec
	UPDATE comments c, forumthread th, forumtopic t
	SET
	th.lastposter=c.users_idusers, t.lastposter=c.users_idusers,
	th.lastaddition=c.written, t.lastaddition=c.written,
	t.comments=IF(th.comments IS NULL, 0, t.comments+1),
	t.threads=IF(th.comments IS NULL, IF(t.threads IS NULL, 1, t.threads+1), t.threads),
	th.comments=IF(th.comments IS NULL, 0, th.comments+1),
	th.firstpost=IF(th.firstpost=0, c.idcomments, th.firstpost)
	WHERE c.idcomments=?;
	*/
	if err := queries.RecalculateForumThreadByIdMetaData(r.Context(), int32(threadId)); err != nil {
		log.Printf("Error: recalculateForumThreadByIdMetaData: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err := queries.RebuildForumTopicByIdMetaColumns(r.Context(), int32(topicId)); err != nil {
		log.Printf("Error: rebuildForumTopicByIdMetaColumns: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	wordIds, done := SearchWordIdsFromText(w, r, text, queries)
	if done {
		return
	}

	if InsertWordsToForumSearch(w, r, wordIds, queries, cid) {
		return
	}

	// TODO listUsersSubscribedToThread

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}

func forumTopicThreadReplyCancelPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])
	threadId, _ := strconv.Atoi(vars["thread"])

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicId, threadId)

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}
