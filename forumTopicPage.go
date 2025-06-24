package goa4web

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core"
	"github.com/gorilla/mux"
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
		Threads                 []*GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow
		Categories              []*ForumcategoryPlus
		Category                *ForumcategoryPlus
		CopyDataToSubCategories func(rootCategory *ForumcategoryPlus) *Data
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
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

	categoryRows, err := queries.GetAllForumCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllForumCategories Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}
	topicRow, err := queries.GetForumTopicByIdForUser(r.Context(), GetForumTopicByIdForUserParams{
		UsersIdusers: uid,
		Idforumtopic: int32(topicId),
	})
	if err != nil {
		log.Printf("showTableTopics Error: %s", err)
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

	threadRows, err := queries.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText(r.Context(), GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams{
		UsersIdusers:           uid,
		ForumtopicIdforumtopic: int32(topicId),
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
	data.Threads = threadRows

	CustomForumIndex(data.CoreData, r)

	if err := renderTemplate(w, r, "topicsPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
