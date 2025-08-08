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

func TopicsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Admin                   bool
		Back                    bool
		Subscribed              bool
		Topic                   *ForumtopicPlus
		Threads                 []*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow
		Categories              []*ForumcategoryPlus
		Category                *ForumcategoryPlus
		CopyDataToSubCategories func(rootCategory *ForumcategoryPlus) *Data
	}

	if _, ok := core.GetSessionOrFail(w, r); !ok {
		return
	}
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	data := &Data{
		Admin: cd.IsAdmin() && cd.IsAdminMode(),
	}

	copyDataToSubCategories := func(rootCategory *ForumcategoryPlus) *Data {
		d := *data
		d.Categories = []*ForumcategoryPlus{}
		d.Category = rootCategory
		d.Back = false
		return &d
	}
	data.CopyDataToSubCategories = copyDataToSubCategories

	categoryRows, err := cd.ForumCategories()
	if err != nil {
		log.Printf("getAllForumCategories Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	topicRow, err := cd.ForumTopicByID(int32(topicId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
		} else {
			log.Printf("showTableTopics Error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		}
		return
	}
	cd.PageTitle = fmt.Sprintf("Forum - %s", topicRow.Title.String)
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
		Edit:                         false,
	}

	categoryTree := NewCategoryTree(categoryRows, []*ForumtopicPlus{data.Topic})
	if category, ok := categoryTree.CategoryLookup[topicRow.ForumcategoryIdforumcategory]; ok {
		category.Topics = []*ForumtopicPlus{
			data.Topic,
		}
		data.Categories = []*ForumcategoryPlus{
			category,
		}
		data.Category = category
	}

	threadRows, err := cd.ForumThreads(int32(topicId))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Error: ForumThreads: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	data.Threads = threadRows

	if subscribedToTopic(cd, topicRow.Idforumtopic) {
		data.Subscribed = true
	}

	handlers.TemplateHandler(w, r, "topicsPage.gohtml", data)
}
