package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/exp/slices"
	"log"
	"net/http"
	"strconv"
)

func writingsArticlePage(w http.ResponseWriter, r *http.Request) {
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
		Writing             *GetWritingByIdForUserDescendingByPublishedDateRow
		CanEdit             bool
		IsAuthor            bool
		CanReply            bool
		UserId              int32
		Languages           []*Language
		SelectedLanguageId  int
		Thread              *GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsRow
		Comments            []*CommentPlus
		IsReplyable         bool
		IsAdmin             bool
		Categories          []*Writingcategory
		CategoryId          int32
		Offset              int32
		CategoryBreadcrumbs []*Writingcategory
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		CanReply: true,  // TODO
		CanEdit:  false, // TODO
	}

	vars := mux.Vars(r)
	articleId, _ := strconv.Atoi(vars["article"])

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)
	data.UserId = uid
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	writing, err := queries.GetWritingByIdForUserDescendingByPublishedDate(r.Context(), GetWritingByIdForUserDescendingByPublishedDateParams{
		Userid:    uid,
		Idwriting: int32(articleId),
	})
	if err != nil {
		log.Printf("getWritingByIdForUserDescendingByPublishedDate Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.Writing = writing
	data.IsAuthor = writing.UsersIdusers == uid
	data.CategoryId = writing.WritingcategoryIdwritingcategory

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		log.Printf("FetchLanguages Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	commentRows, err := queries.GetCommentsByThreadIdForUser(r.Context(), GetCommentsByThreadIdForUserParams{
		UsersIdusers:             uid,
		ForumthreadIdforumthread: writing.ForumthreadIdforumthread,
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

	threadRow, err := queries.GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissions(r.Context(), GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsParams{
		UsersIdusers:  uid,
		Idforumthread: writing.ForumthreadIdforumthread,
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

	categoryRows, err := queries.FetchAllCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllWritingCategories Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	categoryMap := map[int32]*Writingcategory{}
	for _, cat := range categoryRows {
		categoryMap[cat.Idwritingcategory] = cat
		if cat.WritingcategoryIdwritingcategory == data.CategoryId {
			data.Categories = append(data.Categories, cat)
		}
	}
	for cid := data.CategoryId; len(data.CategoryBreadcrumbs) < len(categoryRows); {
		cat, ok := categoryMap[cid]
		if ok {
			data.CategoryBreadcrumbs = append(data.CategoryBreadcrumbs, cat)
			cid = cat.WritingcategoryIdwritingcategory
		} else {
			break
		}
	}
	slices.Reverse(data.CategoryBreadcrumbs)

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

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

	CustomWritingsIndex(data.CoreData, r)

	renderTemplate(w, r, "writingsArticlePage.gohtml", data)
}

func writingsArticleReplyActionPage(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)

	vars := mux.Vars(r)
	aid, err := strconv.Atoi(vars["post"])

	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if aid == 0 {
		log.Printf("Error: no bid")
		http.Redirect(w, r, "?error="+"No bid", http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	uid, _ := session.Values["UID"].(int32)

	post, err := queries.GetWritingByIdForUserDescendingByPublishedDate(r.Context(), GetWritingByIdForUserDescendingByPublishedDateParams{
		Userid:    uid,
		Idwriting: int32(aid),
	})
	if err != nil {
		log.Printf("getArticlePost Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var pthid int32 = post.ForumthreadIdforumthread
	pt, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{
		String: WritingTopicName,
		Valid:  true,
	})
	var ptid int32
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.CreateForumTopic(r.Context(), CreateForumTopicParams{
			ForumcategoryIdforumcategory: 0,
			Title: sql.NullString{
				String: WritingTopicName,
				Valid:  true,
			},
			Description: sql.NullString{
				String: WritingTopicDescription,
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
	} else {
		ptid = pt.Idforumtopic
	}
	if pthid == 0 {
		pthidi, err := queries.MakeThread(r.Context(), ptid)
		if err != nil {
			log.Printf("Error: makeThread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		pthid = int32(pthidi)
		if err := queries.AssignWritingThisThreadId(r.Context(), AssignWritingThisThreadIdParams{
			ForumthreadIdforumthread: pthid,
			Idwriting:                int32(aid),
		}); err != nil {
			log.Printf("Error: assign_article_to_thread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))

	endUrl := fmt.Sprintf("/article/%d", aid)

	provider := getEmailProvider()

	if rows, err := queries.ListUsersSubscribedToThread(r.Context(), ListUsersSubscribedToThreadParams{
		ForumthreadIdforumthread: pthid,
		Idusers:                  uid,
	}); err != nil {
		log.Printf("Error: listUsersSubscribedToThread: %s", err)
	} else if provider != nil {
		for _, row := range rows {
			if err := notifyChange(r.Context(), provider, row.Username.String, endUrl); err != nil {
				log.Printf("Error: notifyChange: %s", err)
			}
		}
	}

	// TODO
	//if rows, err := queries.SomethingNotifyArticle(r.Context(), SomethingNotifyArticlesParams{
	//	Idusers: uid,
	//	Idarticles: int32(bid),
	//}); err != nil {
	//	log.Printf("Error: listUsersSubscribedToThread: %s", err)
	//} else {
	//	for _, row := range rows {
	//		if err := notifyChange(r.Context(), getEmailProvider(), row.String, endUrl); err != nil {
	//			log.Printf("Error: notifyChange: %s", err)
	//
	//		}
	//	}
	//}

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

	if err := PostUpdate(r.Context(), queries, pthid, ptid); err != nil {
		log.Printf("Error: postUpdate: %s", err)
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
