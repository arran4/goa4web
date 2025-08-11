package forum

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"

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
	_, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	categoryId, _ := strconv.Atoi(vars["category"])

	data := &Data{
		Admin: cd.IsAdmin() && cd.IsAdminMode(),
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
	rows, err := cd.ForumTopics(int32(categoryId))
	if err != nil {
		log.Printf("ForumTopics Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	for _, row := range rows {
		topicRows = append(topicRows, &ForumtopicPlus{
			ID:                 row.ID,
			LastAuthorID:       row.LastAuthorID,
			CategoryID:         row.CategoryID,
			Title:              row.Title,
			Description:        row.Description,
			Threads:            row.Threads,
			Comments:           row.Comments,
			Lastaddition:       row.Lastaddition,
			LastAuthorUsername: row.LastAuthorUsername,
		})
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
	// Administrative actions moved to the admin portal.
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
