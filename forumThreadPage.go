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

func forumThreadPage(w http.ResponseWriter, r *http.Request) {
	type CommentPlus struct {
		*User_get_all_comments_for_threadRow
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
		Category            *ForumcategoryPlus
		Topic               *ForumtopicPlus
		Thread              *User_get_threadRow
		Comments            []*CommentPlus
		Offset              int
		IsReplyable         bool
		Text                string
		Languages           []*Language
		SelectedLanguageId  int
		CategoryBreadcrumbs []*ForumcategoryPlus
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	data := Data{
		CoreData:           r.Context().Value(ContextValues("coreData")).(*CoreData),
		Offset:             offset,
		IsReplyable:        true,
		SelectedLanguageId: 1,
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	vars := mux.Vars(r)
	//topicId, _ := strconv.Atoi(vars["topic"])
	threadId, _ := strconv.Atoi(vars["thread"])
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	commentRows, err := queries.User_get_all_comments_for_thread(r.Context(), User_get_all_comments_for_threadParams{
		UsersIdusers:             uid,
		ForumthreadIdforumthread: int32(threadId),
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
		Idforumthread: int32(threadId),
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

	topicRow, err := queries.User_get_topic(r.Context(), User_get_topicParams{
		UsersIdusers: uid,
		Idforumtopic: int32(threadRow.ForumtopicIdforumtopic),
	})
	if err != nil {
		log.Printf("user_get_topic Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	data.Topic = &ForumtopicPlus{
		Idforumtopic:                 topicRow.Idforumtopic,
		Lastposter:                   topicRow.Lastposter,
		ForumcategoryIdforumcategory: topicRow.ForumcategoryIdforumcategory,
		Title:                        topicRow.Title,
		Description:                  topicRow.Description,
		Threads:                      topicRow.Threads,
		Comments:                     topicRow.Comments,
		Lastaddition:                 topicRow.Lastaddition,
		Lastposterusername:           topicRow.Lastposterusername,
		Seelevel:                     topicRow.Seelevel,
		Level:                        topicRow.Level,
		Edit:                         false,
	}

	commentIdString := r.URL.Query().Get("comment")
	commentId, _ := strconv.Atoi(commentIdString)
	for i, row := range commentRows {
		editUrl := ""
		editSaveUrl := ""
		if uid == row.UsersIdusers {
			editUrl = fmt.Sprintf("/forum/topic/%d/thread/%d?comment=%d#edit", topicRow.Idforumtopic, threadId, row.Idcomments)
			editSaveUrl = fmt.Sprintf("/forum/topic/%d/thread/%d/comment/%d", topicRow.Idforumtopic, threadId, row.Idcomments)
			if commentId != 0 && int32(commentId) == row.Idcomments {
				data.IsReplyable = false
			}
		}

		data.Comments = append(data.Comments, &CommentPlus{
			User_get_all_comments_for_threadRow: row,
			ShowReply:                           true,
			EditUrl:                             editUrl,
			EditSaveUrl:                         editSaveUrl,
			Editing:                             commentId != 0 && int32(commentId) == row.Idcomments,
			Offset:                              i + offset,
			Languages:                           nil,
			SelectedLanguageId:                  0,
		})
	}

	data.Thread = threadRow
	data.Topic = &ForumtopicPlus{
		Idforumtopic:                 topicRow.Idforumtopic,
		Lastposter:                   topicRow.Lastposter,
		ForumcategoryIdforumcategory: topicRow.ForumcategoryIdforumcategory,
		Title:                        topicRow.Title,
		Description:                  topicRow.Description,
		Threads:                      topicRow.Threads,
		Comments:                     topicRow.Comments,
		Lastaddition:                 topicRow.Lastaddition,
		Lastposterusername:           topicRow.Lastposterusername,
		Seelevel:                     topicRow.Seelevel,
		Level:                        topicRow.Level,
		Edit:                         false,
	}

	categoryRows, err := queries.ForumCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("forumCategories Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	categoryTree := NewCategoryTree(categoryRows, []*ForumtopicPlus{data.Topic})
	data.CategoryBreadcrumbs = categoryTree.CategoryRoots(int32(topicRow.ForumcategoryIdforumcategory))

	replyType := r.URL.Query().Get("type")
	if commentIdString != "" {
		comment, err := queries.User_get_comment(r.Context(), User_get_commentParams{
			UsersIdusers: uid,
			Idcomments:   int32(commentId),
		})
		if err != nil {
			log.Printf("user_get_comment Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		switch replyType {
		case "full":
			data.Text = processCommentFullQuote(comment.Username.String, comment.Text.String)
		default:
			data.Text = processCommentQuote(comment.Username.String, comment.Text.String)
		}
	}

	CustomBlogIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "forumThreadPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
