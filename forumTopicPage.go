package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"strconv"
)

func forumTopicsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		CategoryBreadcrumbs []*ForumcategoryPlus
		Admin               bool
		Back                bool
		Topic               *ForumtopicPlus
		Threads             []*get_user_threads_for_topicRow
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])

	data := &Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Admin:    true,
	}

	categoryRows, err := queries.forumCategories(r.Context())
	if err != nil {
		log.Printf("forumCategories Error: %s", err)
		http.Redirect(w, r, "?error="+"No bid", http.StatusTemporaryRedirect)
		return
	}
	topicRow, err := queries.get_user_topic(r.Context(), get_user_topicParams{
		UsersIdusers: uid,
		Idforumtopic: int32(topicId),
	})
	if err != nil {
		log.Printf("showTableTopics Error: %s", err)
		http.Redirect(w, r, "?error="+"No bid", http.StatusTemporaryRedirect)
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

	categoryTree := NewCategoryTree(categoryRows, []*ForumtopicPlus{{
		Idforumtopic:                 topicRow.Idforumtopic,
		Lastposter:                   topicRow.Lastposter,
		ForumcategoryIdforumcategory: topicRow.ForumcategoryIdforumcategory,
		Title:                        topicRow.Title,
		Description:                  topicRow.Description,
		Threads:                      topicRow.Threads,
		Comments:                     topicRow.Comments,
		Lastaddition:                 topicRow.Lastaddition,
	}})
	data.CategoryBreadcrumbs = categoryTree.CategoryRoots(int32(topicRow.ForumcategoryIdforumcategory))

	threadRows, err := queries.get_user_threads_for_topic(r.Context(), get_user_threads_for_topicParams{
		UsersIdusers:           uid,
		ForumtopicIdforumtopic: int32(topicId),
	})
	if err != nil {
		log.Printf("forumCategories Error: %s", err)
		http.Redirect(w, r, "?error="+"No bid", http.StatusTemporaryRedirect)
		return
	}
	data.Threads = threadRows

	CustomForumIndex(data.CoreData, r)

	if err := compiledTemplates.ExecuteTemplate(w, "forumTopicsPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
