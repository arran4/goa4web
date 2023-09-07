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

func linkerCommentsPage(w http.ResponseWriter, r *http.Request) {
	type CommentPlus struct {
		*GetCommentsByThreadIdForUserRow
		ShowReply          bool
		EditUrl            string
		Editing            bool
		Offset             int
		Languages          []*Language
		SelectedLanguageId int32
		EditSaveUrl        string
	}
	type Data struct {
		*CoreData
		Link               *GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow
		CanReply           bool
		Languages          []*Language
		Comments           []*CommentPlus
		SelectedLanguageId int
		IsReplyable        bool
		Offset             int
		Text               string
		CanEdit            bool
		UserId             int32
		Thread             *User_get_threadRow
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		CanReply: true,  // TODO
		CanEdit:  false, // TODO
		Offset:   offset,
	}

	vars := mux.Vars(r)
	linkId, _ := strconv.Atoi(vars["link"])
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)
	data.UserId = uid

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending(r.Context(), int32(linkId))
	if err != nil {
		log.Printf("getLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.Link = link

	commentRows, err := queries.GetCommentsByThreadIdForUser(r.Context(), GetCommentsByThreadIdForUserParams{
		UsersIdusers:             uid,
		ForumthreadIdforumthread: link.ForumthreadIdforumthread,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getBlogEntryForUserById_comments Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	threadRow, err := queries.User_get_thread(r.Context(), User_get_threadParams{
		UsersIdusers:  uid,
		Idforumthread: link.ForumthreadIdforumthread,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("Error: user_get_thread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))

	commentIdString := r.URL.Query().Get("comment")
	commentId, _ := strconv.Atoi(commentIdString)
	for i, row := range commentRows {
		editUrl := ""
		editSaveUrl := ""
		if uid == row.UsersIdusers {
			// TODO
			//editUrl = fmt.Sprintf("/forum/topic/%d/thread/%d?comment=%d#edit", topicRow.Idforumtopic, threadId, row.Idcomments)
			//editSaveUrl = fmt.Sprintf("/forum/topic/%d/thread/%d/comment/%d", topicRow.Idforumtopic, threadId, row.Idcomments)
			if commentId != 0 && int32(commentId) == row.Idcomments {
				data.IsReplyable = false
			}
		}

		data.Comments = append(data.Comments, &CommentPlus{
			GetCommentsByThreadIdForUserRow: row,
			ShowReply:                       true,
			EditUrl:                         editUrl,
			EditSaveUrl:                     editSaveUrl,
			Editing:                         commentId != 0 && int32(commentId) == row.Idcomments,
			Offset:                          i + offset,
			Languages:                       nil,
			SelectedLanguageId:              0,
		})
	}

	data.Thread = threadRow

	CustomLinkerIndex(data.CoreData, r)
	if err := getCompiledTemplates().ExecuteTemplate(w, "linkerCommentsPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func linkerCommentsReplyPage(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)

	vars := mux.Vars(r)
	linkId, err := strconv.Atoi(vars["link"])

	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if linkId == 0 {
		log.Printf("Error: no bid")
		http.Redirect(w, r, "?error="+"No bid", http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending(r.Context(), int32(linkId))
	if err != nil {
		log.Printf("getLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var pthid int32 = link.ForumthreadIdforumthread
	ptid, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{
		String: LinkderTopicName,
		Valid:  true,
	})
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.CreateForumTopic(r.Context(), CreateForumTopicParams{
			ForumcategoryIdforumcategory: 0,
			Title: sql.NullString{
				String: LinkderTopicName,
				Valid:  true,
			},
			Description: sql.NullString{
				String: LinkderTopicDescription,
				Valid:  true,
			},
		})
		if err != nil {
			log.Printf("Error: createForumTopic: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		ptid = int32(ptidi)
	} else if err != nil {
		log.Printf("Error: findForumTopicByTitle: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if pthid == 0 {
		pthidi, err := queries.MakeThread(r.Context(), ptid)
		if err != nil {
			log.Printf("Error: makeThread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		pthid = int32(pthidi)
		if err := queries.AssignLinkerThisThreadId(r.Context(), AssignLinkerThisThreadIdParams{
			ForumthreadIdforumthread: pthid,
			Idlinker:                 int32(linkId),
		}); err != nil {
			log.Printf("Error: assignThreadIdToBlogEntry: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	uid, _ := session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/linker/comments/%d", linkId)

	if rows, err := queries.ListUsersSubscribedToThread(r.Context(), ListUsersSubscribedToThreadParams{
		ForumthreadIdforumthread: pthid,
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

	if rows, err := queries.ListUsersSubscribedToLinker(r.Context(), ListUsersSubscribedToLinkerParams{
		Idusers:  uid,
		Idlinker: int32(linkId),
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
		ForumthreadIdforumthread: pthid,
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		log.Printf("Error: createComment: %s", err)
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
	if err := queries.Update_forumthread(r.Context(), pthid); err != nil {
		log.Printf("Error: update_forumthread: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err := queries.RebuildForumTopicByIdMetaColumns(r.Context(), ptid); err != nil {
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

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}
