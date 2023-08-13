package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

func forumPage(w http.ResponseWriter, r *http.Request) {

	type ForumtopicPlus struct {
		*showTableTopicsRow
		Edit bool
	}

	type ForumcategoryPlus struct {
		*Forumcategory
		Categories []*ForumcategoryPlus
		Topics     []*ForumtopicPlus
		Edit       bool
	}

	type Data struct {
		*CoreData
		Categories              []*ForumcategoryPlus
		Admin                   bool
		CopyDataToSubCategories func(rootCategory *ForumcategoryPlus) *Data
		Category                *ForumcategoryPlus
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	data := &Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Admin:    true,
	}

	copyDataToSubCategories := func(rootCategory *ForumcategoryPlus) *Data {
		d := *data
		d.Categories = rootCategory.Categories
		d.Category = rootCategory
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
	categoryParents := map[int32][]*ForumcategoryPlus{}
	categories := map[int32]*ForumcategoryPlus{}
	for _, row := range categoryRows {
		fcp := &ForumcategoryPlus{
			Forumcategory: row,
			Categories:    nil,
			Edit:          false,
		}
		categoryParents[row.ForumcategoryIdforumcategory] = append(categoryParents[row.ForumcategoryIdforumcategory], fcp)
		categories[row.Idforumcategory] = fcp
	}
	for _, row := range topicRows {
		tp := &ForumtopicPlus{
			showTableTopicsRow: row,
		}
		c, ok := categories[row.ForumcategoryIdforumcategory]
		if !ok || c == nil {
			continue
		}
		c.Topics = append(c.Topics, tp)
	}
	for parentId, children := range categoryParents {
		parent, ok := categories[parentId]
		if !ok {
			continue
		}
		parent.Categories = children
	}
	data.Categories = categoryParents[0]

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
