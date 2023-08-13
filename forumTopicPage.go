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
		CategoryBreadcrumbs     []*ForumcategoryPlus
		Admin                   bool
		Back                    bool
		Topic                   *ForumtopicPlus
		Threads                 []*user_get_threads_for_topicRow
		Categories              []*ForumcategoryPlus
		Category                *ForumcategoryPlus
		CopyDataToSubCategories func(rootCategory *ForumcategoryPlus) *Data
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

	copyDataToSubCategories := func(rootCategory *ForumcategoryPlus) *Data {
		d := *data
		d.Categories = []*ForumcategoryPlus{}
		d.Category = rootCategory
		d.Back = false
		return &d
	}
	data.CopyDataToSubCategories = copyDataToSubCategories

	categoryRows, err := queries.forumCategories(r.Context())
	if err != nil {
		log.Printf("forumCategories Error: %s", err)
		http.Redirect(w, r, "?error="+"No bid", http.StatusTemporaryRedirect)
		return
	}
	topicRow, err := queries.user_get_topic(r.Context(), user_get_topicParams{
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

	categoryTree := NewCategoryTree(categoryRows, []*ForumtopicPlus{data.Topic})
	data.CategoryBreadcrumbs = categoryTree.CategoryRoots(int32(topicRow.ForumcategoryIdforumcategory))
	if category, ok := categoryTree.CategoryLookup[topicRow.ForumcategoryIdforumcategory]; ok {
		category.Topics = []*ForumtopicPlus{
			data.Topic,
		}
		data.Categories = []*ForumcategoryPlus{
			category,
		}
	}

	threadRows, err := queries.user_get_threads_for_topic(r.Context(), user_get_threads_for_topicParams{
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
