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

func forumPage(w http.ResponseWriter, r *http.Request) {

	type Data struct {
		*CoreData
		Categories              []*ForumcategoryPlus
		CategoryBreadcrumbs     []*ForumcategoryPlus
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
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("forumCategories Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}
	var topicRows []*ForumtopicPlus
	if categoryId == 0 {
		rows, err := queries.get_all_user_topics(r.Context(), uid)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
			default:
				log.Printf("showTableTopics Error: %s", err)
				http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
				return
			}
		}
		for _, row := range rows {
			topicRows = append(topicRows, &ForumtopicPlus{
				Idforumtopic:                 row.Idforumtopic,
				Lastposter:                   row.Lastposter,
				ForumcategoryIdforumcategory: row.ForumcategoryIdforumcategory,
				Title:                        row.Title,
				Description:                  row.Description,
				Threads:                      row.Threads,
				Comments:                     row.Comments,
				Lastaddition:                 row.Lastaddition,
			})
		}
	} else {
		rows, err := queries.get_all_user_topics_for_category(r.Context(), get_all_user_topics_for_categoryParams{
			UsersIdusers:                 uid,
			ForumcategoryIdforumcategory: int32(categoryId),
		})
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
			default:
				log.Printf("showTableTopics Error: %s", err)
				http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
				return
			}
		}
		for _, row := range rows {
			topicRows = append(topicRows, &ForumtopicPlus{
				Idforumtopic:                 row.Idforumtopic,
				Lastposter:                   row.Lastposter,
				ForumcategoryIdforumcategory: row.ForumcategoryIdforumcategory,
				Title:                        row.Title,
				Description:                  row.Description,
				Threads:                      row.Threads,
				Comments:                     row.Comments,
				Lastaddition:                 row.Lastaddition,
			})
		}
	}

	categoryTree := NewCategoryTree(categoryRows, topicRows)

	if categoryId == 0 {
		data.Categories = categoryTree.CategoryChildrenLookup[int32(categoryId)]
	} else if cat, ok := categoryTree.CategoryLookup[int32(categoryId)]; ok && cat != nil {
		data.Categories = []*ForumcategoryPlus{cat}
		data.Category = cat
		data.CategoryBreadcrumbs = categoryTree.CategoryRoots(int32(categoryId))
		data.Back = true
	}

	CustomForumIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "forumPage.gohtml", data); err != nil {
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
				Name: "Administer topics",
				Link: "/forum/admin/topics",
			},
		)
		data.CustomIndexItems = append(data.CustomIndexItems,
			IndexItem{
				Name: "Administer users",
				Link: "/forum/admin/users",
			},
		)
		data.CustomIndexItems = append(data.CustomIndexItems,
			IndexItem{
				Name: "Administer topic restrictions",
				Link: "/forum/admin/restrictions/topics",
			},
		)
		data.CustomIndexItems = append(data.CustomIndexItems,
			IndexItem{
				Name: "Administer user restrictions",
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
