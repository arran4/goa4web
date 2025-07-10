package forum

import (
	"database/sql"
	"errors"
	"fmt"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
	"github.com/gorilla/mux"
)

func Page(w http.ResponseWriter, r *http.Request) {

	type Data struct {
		*CoreData
		Categories              []*ForumcategoryPlus
		CategoryBreadcrumbs     []*ForumcategoryPlus
		Admin                   bool
		CopyDataToSubCategories func(rootCategory *ForumcategoryPlus) *Data
		Category                *ForumcategoryPlus
		Back                    bool
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	categoryId, _ := strconv.Atoi(vars["category"])

	data := &Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
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
	var topicRows []*ForumtopicPlus
	if categoryId == 0 {
		rows, err := queries.GetAllForumTopicsForUser(r.Context(), uid)
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
		rows, err := queries.GetAllForumTopicsByCategoryIdForUserWithLastPosterName(r.Context(), db.GetAllForumTopicsByCategoryIdForUserWithLastPosterNameParams{
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

	if err := templates.RenderTemplate(w, "forumPage", data, corecommon.NewFuncs(r)); err != nil {
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
	if data.FeedsEnabled && topicId != "" && threadId == "" {
		data.RSSFeedUrl = fmt.Sprintf("/forum/topic/%s.rss", topicId)
		data.AtomFeedUrl = fmt.Sprintf("/forum/topic/%s.atom", topicId)
		data.CustomIndexItems = append(data.CustomIndexItems,
			IndexItem{Name: "Atom Feed", Link: data.AtomFeedUrl},
			IndexItem{Name: "RSS Feed", Link: data.RSSFeedUrl},
		)
	}
	userHasAdmin := data.HasRole("administrator") && data.AdminMode
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
	if threadId != "" && topicId != "" {
		if tid, err := strconv.Atoi(topicId); err == nil && CanReply(r, int32(tid)) {
			data.CustomIndexItems = append(data.CustomIndexItems,
				IndexItem{
					Name: "Write Reply",
					Link: fmt.Sprintf("/forum/topic/%s/thread/%s/reply", topicId, threadId),
				},
			)
		}
	}
	if categoryId != "" && topicId != "" {
		if tid, err := strconv.Atoi(topicId); err == nil && CanCreateThread(r, int32(tid)) {
			data.CustomIndexItems = append(data.CustomIndexItems,
				IndexItem{
					Name: "Create Thread",
					Link: fmt.Sprintf("/forum/topic/%s/new", topicId),
				},
			)
		}
	}
}
