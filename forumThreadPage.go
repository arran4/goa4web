package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"strconv"
)

func forumThreadPage(w http.ResponseWriter, r *http.Request) {
	type CommentPlus struct {
		*user_get_all_comments_for_threadRow
		ShowReply          bool
		EditUrl            string
		Editing            bool
		Offset             int
		Languages          []*Language
		SelectedLanguageId int32
	}
	type Data struct {
		*CoreData
		Category            *ForumcategoryPlus
		Topic               *ForumtopicPlus
		Comments            []*CommentPlus
		Offset              int
		IsReplyable         bool
		Text                string
		Languages           []*Language
		SelectedLanguageId  int
		Thread              *user_get_threadRow
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

	languageRows, err := queries.fetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	vars := mux.Vars(r)
	threadId, _ := strconv.Atoi(vars["thread"])
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	commentRows, err := queries.user_get_all_comments_for_thread(r.Context(), user_get_all_comments_for_threadParams{
		UsersIdusers:             uid,
		ForumthreadIdforumthread: int32(threadId),
	})
	if err != nil {
		log.Printf("show_blog_comments Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	threadRow, err := queries.user_get_thread(r.Context(), user_get_threadParams{
		UsersIdusers:  uid,
		Idforumthread: int32(threadId),
	})
	if err != nil {
		log.Printf("showTableThreads Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	topicRow, err := queries.user_get_topic(r.Context(), user_get_topicParams{
		UsersIdusers: uid,
		Idforumtopic: int32(threadRow.ForumtopicIdforumtopic),
	})
	if err != nil {
		log.Printf("showTableTopics Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	for i, row := range commentRows {
		editUrl := ""
		if uid == row.UsersIdusers {
			editUrl = fmt.Sprintf("/forum/topic/%d/thread/%d/comment/%d/edit#edit", topicRow.Idforumtopic, threadId, row.Idcomments)
		}

		data.Comments = append(data.Comments, &CommentPlus{
			user_get_all_comments_for_threadRow: row,
			ShowReply:                           true,
			EditUrl:                             editUrl,
			Offset:                              i + offset,
			Editing:                             false,
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

	categoryRows, err := queries.forumCategories(r.Context())
	if err != nil {
		log.Printf("forumCategories Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	categoryTree := NewCategoryTree(categoryRows, []*ForumtopicPlus{data.Topic})
	data.CategoryBreadcrumbs = categoryTree.CategoryRoots(int32(topicRow.ForumcategoryIdforumcategory))
	CustomBlogIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "forumThreadPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
