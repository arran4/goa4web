package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"strconv"
)

func forumPage(w http.ResponseWriter, r *http.Request) {

	type Data struct {
		*CoreData
		Categories              []*ForumcategoryPlus
		Admin                   bool
		CopyDataToSubCategories func(rootCategory *ForumcategoryPlus) *Data
		Category                *ForumcategoryPlus
		Back                    bool
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	categoryId, _ := strconv.Atoi(vars["category"])

	data := &Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Admin:    true,
	}

	copyDataToSubCategories := func(rootCategory *ForumcategoryPlus) *Data {
		d := *data
		d.Categories = rootCategory.Categories
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
	topicRows, err := queries.showTableTopics(r.Context(), uid)
	if err != nil {
		log.Printf("showTableTopics Error: %s", err)
		http.Redirect(w, r, "?error="+"No bid", http.StatusTemporaryRedirect)
		return
	}

	categoryTree := NewCategoryTree(categoryRows, topicRows)

	if categoryId == 0 {
		data.Categories = categoryTree.CategoryParentLookup[int32(categoryId)]
	} else if cat, ok := categoryTree.CategoryLookup[int32(categoryId)]; ok && cat != nil {
		data.Categories = []*ForumcategoryPlus{cat}
		data.Category = cat
		data.Back = true
	}

	CustomForumIndex(data.CoreData, r)

	if err := compiledTemplates.ExecuteTemplate(w, "forumPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CustomForumIndex(data *CoreData, r *http.Request) {
	vars := mux.Vars(r)
	threadId := vars["thread"]
	topicId := vars["topic"]
	categoryId := vars["category"]
	userHasAdmin := true // TODO
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems,
			IndexItem{
				Name: "Admin",
				Link: "/forum/admin",
			},
		)
		data.CustomIndexItems = append(data.CustomIndexItems,
			IndexItem{
				Name: "Administer categories",
				Link: "/forum/admin/categories",
			},
		)
		data.CustomIndexItems = append(data.CustomIndexItems,
			IndexItem{
				Name: "Administer create category",
				Link: "/forum/admin/categories/create",
			},
		)
		data.CustomIndexItems = append(data.CustomIndexItems,
			IndexItem{
				Name: "Administer topics",
				Link: "/forum/admin/topics",
			},
		)
		data.CustomIndexItems = append(data.CustomIndexItems,
			IndexItem{
				Name: "Administer create topic",
				Link: "/forum/admin/topics/create",
			},
		)
		data.CustomIndexItems = append(data.CustomIndexItems,
			IndexItem{
				Name: "Administer users",
				Link: "/forum/admin/user",
			},
		)
		data.CustomIndexItems = append(data.CustomIndexItems,
			IndexItem{
				Name: "Administer restrictions",
				Link: "/forum/admin/restrictions",
			},
		)
		data.CustomIndexItems = append(data.CustomIndexItems,
			IndexItem{
				Name: "Administer restrictions",
				Link: "/forum/admin/restrictions/users",
			},
		)
	}
	if threadId != "" && topicId != "" { // TODO Permissions system
		data.CustomIndexItems = append(data.CustomIndexItems,
			IndexItem{
				Name: "Write Reply",
				Link: fmt.Sprintf("/forum/topic/%s/thread/%s/reply", topicId, threadId),
			},
		)
	}
	if categoryId != "" && topicId != "" { // TODO Permissions system
		data.CustomIndexItems = append(data.CustomIndexItems,
			IndexItem{
				Name: "Create Thread",
				Link: fmt.Sprintf("/forum/topic/%s/new", topicId),
			},
		)
	}
}
