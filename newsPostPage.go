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

type NewsPost struct {
	ShowReply bool
	ShowEdit  bool
	// TODO or (eq .Level "authWriter") (and (ge .Level "authModerator") (le .Level "authAdministrator"))
}

func newsPostPage(w http.ResponseWriter, r *http.Request) {
	type CommentPlus struct {
		*User_get_all_comments_for_threadRow
		ShowReply          bool
		EditUrl            string
		Editing            bool
		Offset             int
		Languages          []*Language
		SelectedLanguageId int
		EditSaveUrl        string
	}
	type Post struct {
		*GetNewsPostRow
		ShowReply bool
		ShowEdit  bool
		Editing   bool
	}
	type Data struct {
		*CoreData
		Post               *Post
		Languages          []*Language
		SelectedLanguageId int32
		Topic              *Forumtopic
		Comments           []*CommentPlus
		Offset             int
		IsReplying         bool
		IsReplyable        bool
		Thread             *User_get_threadRow
		ReplyText          string
	}

	data := Data{
		CoreData:    r.Context().Value(ContextValues("coreData")).(*CoreData),
		IsReplying:  r.URL.Query().Has("comment"),
		IsReplyable: true,
	}
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	post, err := queries.GetNewsPost(r.Context(), int32(pid))
	if err != nil {
		log.Printf("getNewsPost Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	editingId, _ := strconv.Atoi(r.URL.Query().Get("edit"))
	replyType := r.URL.Query().Get("type")

	commentRows, err := queries.User_get_all_comments_for_thread(r.Context(), User_get_all_comments_for_threadParams{
		UsersIdusers:             uid,
		ForumthreadIdforumthread: int32(post.ForumthreadIdforumthread),
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("show_blog_comments Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	threadRow, err := queries.User_get_thread(r.Context(), User_get_threadParams{
		UsersIdusers:  uid,
		Idforumthread: int32(post.ForumthreadIdforumthread),
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

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	commentIdString := r.URL.Query().Get("comment")
	commentId, _ := strconv.Atoi(commentIdString)

	editCommentIdString := r.URL.Query().Get("editComment")
	editCommentId, _ := strconv.Atoi(editCommentIdString)
	for i, row := range commentRows {
		editUrl := fmt.Sprintf("?edit=%d", row.Idcomments)
		editSaveUrl := "?"
		if uid == row.UsersIdusers {
			// TODO
			//editUrl = fmt.Sprintf("/forum/topic/%d/thread/%d?comment=%d#edit", topicRow.Idforumtopic, threadId, row.Idcomments)
			//editSaveUrl = fmt.Sprintf("/forum/topic/%d/thread/%d/comment/%d", topicRow.Idforumtopic, threadId, row.Idcomments)
			if commentId != 0 && int32(commentId) == row.Idcomments {
				data.IsReplyable = false
			}
		}

		if int32(commentId) == row.Idcomments {
			switch replyType {
			case "full":
				data.ReplyText = processCommentFullQuote(row.Posterusername.String, row.Text.String)
			default:
				data.ReplyText = processCommentQuote(row.Posterusername.String, row.Text.String)
			}
		}

		data.Comments = append(data.Comments, &CommentPlus{
			User_get_all_comments_for_threadRow: row,
			ShowReply:                           true,
			EditUrl:                             editUrl,
			EditSaveUrl:                         editSaveUrl,
			Editing:                             editCommentId != 0 && int32(editCommentId) == row.Idcomments,
			Offset:                              i + offset,
			Languages:                           nil,
			SelectedLanguageId:                  0,
		})
	}

	data.Thread = threadRow
	data.Post = &Post{
		GetNewsPostRow: post,
		ShowReply:      true, // TODO
		ShowEdit:       true, // TODO
		Editing:        editingId == int(post.Idsitenews),
	}

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	CustomNewsIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "newsPostPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func newsPostReplyActionPage(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)

	vars := mux.Vars(r)
	pid, err := strconv.Atoi(vars["post"])

	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if pid == 0 {
		log.Printf("Error: no bid")
		http.Redirect(w, r, "?error="+"No bid", http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	post, err := queries.GetNewsPost(r.Context(), int32(pid))
	if err != nil {
		log.Printf("getNewsPost Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var pthid = post.ForumthreadIdforumthread
	ptid, err := queries.FindForumTopicByName(r.Context(), sql.NullString{
		String: NewsTopicName,
		Valid:  true,
	})
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.MakeTopic(r.Context(), MakeTopicParams{
			ForumcategoryIdforumcategory: 0,
			Title: sql.NullString{
				String: NewsTopicName,
				Valid:  true,
			},
			Description: sql.NullString{
				String: NewsTopicDescription,
				Valid:  true,
			},
		})
		if err != nil {
			log.Printf("Error: makeTopic: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		ptid = int32(ptidi)
	} else if err != nil {
		log.Printf("Error: findForumTopicByName: %s", err)
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
		if err := queries.AssignNewsThisThreadId(r.Context(), AssignNewsThisThreadIdParams{
			ForumthreadIdforumthread: pthid,
			Idsitenews:               int32(pid),
		}); err != nil {
			log.Printf("Error: assign_news_to_thread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	uid, _ := session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/news/news/%d", pid)

	if rows, err := queries.ThreadNotify(r.Context(), ThreadNotifyParams{
		ForumthreadIdforumthread: pthid,
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

	// TODO
	//if rows, err := queries.SomethingNotifyNews(r.Context(), somethingNotifyNewssParams{
	//	Idusers: uid,
	//	Idnewss: int32(bid),
	//}); err != nil {
	//	log.Printf("Error: threadNotify: %s", err)
	//} else {
	//	for _, row := range rows {
	//		if err := notifyChange(r.Context(), getEmailProvider(), row.String, endUrl); err != nil {
	//			log.Printf("Error: notifyChange: %s", err)
	//
	//		}
	//	}
	//}

	cid, err := queries.MakePost(r.Context(), MakePostParams{
		LanguageIdlanguage:       int32(languageId),
		UsersIdusers:             uid,
		ForumthreadIdforumthread: pthid,
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		log.Printf("Error: makePost: %s", err)
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

	if err := queries.Update_forumtopic(r.Context(), ptid); err != nil {
		log.Printf("Error: update_forumtopic: %s", err)
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

	taskDoneAutoRefreshPage(w, r)
}

func newsPostEditActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO verify field names
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	postId, _ := strconv.Atoi(vars["post"])

	err = queries.EditNewsPost(r.Context(), EditNewsPostParams{
		Idsitenews:         int32(postId),
		LanguageIdlanguage: int32(languageId),
		News: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	taskDoneAutoRefreshPage(w, r)
}

func newsPostNewActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO verify field names
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("text")
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	err = queries.WriteNewsPost(r.Context(), WriteNewsPostParams{
		LanguageIdlanguage: int32(languageId),
		News: sql.NullString{
			String: text,
			Valid:  true,
		},
		UsersIdusers: uid,
	})
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	taskDoneAutoRefreshPage(w, r)
}
