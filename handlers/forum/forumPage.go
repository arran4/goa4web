package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/core"
	"github.com/gorilla/mux"
)

func Page(w http.ResponseWriter, r *http.Request) {

	type Data struct {
		Categories              []*ForumcategoryPlus
		Admin                   bool
		CopyDataToSubCategories func(rootCategory *ForumcategoryPlus) *Data
		Category                *ForumcategoryPlus
		Back                    bool
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum"
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	categoryId, _ := strconv.Atoi(vars["category"])

	data := &Data{
		Admin: cd.CanEditAny(),
	}

	copyDataToSubCategories := func(rootCategory *ForumcategoryPlus) *Data {
		d := *data
		d.Categories = rootCategory.Categories
		d.Category = rootCategory
		d.Back = false
		return &d
	}
	data.CopyDataToSubCategories = copyDataToSubCategories

	categoryRows, err := cd.ForumCategories()
	if err != nil {
		log.Printf("getAllForumCategories Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	var topicRows []*ForumtopicPlus
	if categoryId == 0 {
		rows, err := queries.GetAllForumTopicsForUser(r.Context(), db.GetAllForumTopicsForUserParams{
			ViewerID:      uid,
			ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
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
	} else {
		rows, err := queries.GetAllForumTopicsByCategoryIdForUserWithLastPosterName(r.Context(), db.GetAllForumTopicsByCategoryIdForUserWithLastPosterNameParams{
			ViewerID:      uid,
			CategoryID:    int32(categoryId),
			ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
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
		data.Back = true
	}

	handlers.TemplateHandler(w, r, "forumPage", data)
}

func CustomForumIndex(data *common.CoreData, r *http.Request) {
	vars := mux.Vars(r)
	threadId := vars["thread"]
	topicId := vars["topic"]
	categoryId := vars["category"]
	data.CustomIndexItems = []common.IndexItem{}
	if data.FeedsEnabled && topicId != "" && threadId == "" {
		data.RSSFeedURL = fmt.Sprintf("/forum/topic/%s.rss", topicId)
		data.AtomFeedURL = fmt.Sprintf("/forum/topic/%s.atom", topicId)
		data.CustomIndexItems = append(data.CustomIndexItems,
			common.IndexItem{Name: "Atom Feed", Link: data.AtomFeedURL},
			common.IndexItem{Name: "RSS Feed", Link: data.RSSFeedURL},
		)
	}
	userHasAdmin := data.HasRole("administrator") && data.AdminMode
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems,
			common.IndexItem{
				Name: "Admin",
				Link: "/admin/forum",
			},
		)
		data.CustomIndexItems = append(data.CustomIndexItems,
			common.IndexItem{
				Name: "Administer categories",
				Link: "/admin/forum/categories",
			},
		)
		data.CustomIndexItems = append(data.CustomIndexItems,
			common.IndexItem{
				Name: "Administer topics",
				Link: "/admin/forum/topics",
			},
		)
		data.CustomIndexItems = append(data.CustomIndexItems,
			common.IndexItem{
				Name: "Administer users",
				Link: "/admin/forum/users",
			},
		)
	}
	if threadId != "" && topicId != "" {
		if tid, err := strconv.Atoi(topicId); err == nil && data.HasGrant("forum", "topic", "reply", int32(tid)) {
			data.CustomIndexItems = append(data.CustomIndexItems,
				common.IndexItem{
					Name: "Write Reply",
					Link: fmt.Sprintf("/forum/topic/%s/thread/%s/reply", topicId, threadId),
				},
			)
		}
	}
	if categoryId != "" && topicId != "" {
		if tid, err := strconv.Atoi(topicId); err == nil && data.HasGrant("forum", "topic", "post", int32(tid)) {
			data.CustomIndexItems = append(data.CustomIndexItems,
				common.IndexItem{
					Name: "Create Thread",
					Link: fmt.Sprintf("/forum/topic/%s/new", topicId),
				},
			)
		}
	}
	if threadId == "" && topicId != "" && data.UserID != 0 {
		tid, err := strconv.Atoi(topicId)
		if err == nil {
			if subscribedToTopic(data, int32(tid)) {
				data.CustomIndexItems = append(data.CustomIndexItems,
					common.IndexItem{
						Name: "Unsubscribe From Topic",
						Link: fmt.Sprintf("/forum/topic/%s/unsubscribe", topicId),
					},
				)
			} else {
				data.CustomIndexItems = append(data.CustomIndexItems,
					common.IndexItem{
						Name: "Subscribe To Topic",
						Link: fmt.Sprintf("/forum/topic/%s/subscribe", topicId),
					},
				)
			}
		}
	}
}
