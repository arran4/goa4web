package forum

import (
	"database/sql"
	"errors"
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

func TopicsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		CategoryBreadcrumbs     []*ForumcategoryPlus
		Admin                   bool
		Back                    bool
		Topic                   *ForumtopicPlus
		Threads                 []*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow
		Categories              []*ForumcategoryPlus
		Category                *ForumcategoryPlus
		CopyDataToSubCategories func(rootCategory *ForumcategoryPlus) *Data
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])

	cd := r.Context().Value(common.KeyCoreData).(*CoreData)
	data := &Data{
		CoreData: cd,
		Admin:    cd.CanEditAny(),
	}

	copyDataToSubCategories := func(rootCategory *ForumcategoryPlus) *Data {
		d := *data
		d.Categories = []*ForumcategoryPlus{}
		d.Category = rootCategory
		d.Back = false
		return &d
	}
	data.CopyDataToSubCategories = copyDataToSubCategories

	categoryRows, err := data.CoreData.ForumCategories()
	if err != nil {
		log.Printf("getAllForumCategories Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	topicRow, err := data.CoreData.ForumTopicByID(int32(topicId))
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

	threadRows, err := cd.ForumThreads(int32(topicId))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Error: ForumThreads: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	data.Threads = threadRows

	if err := templates.RenderTemplate(w, "topicsPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
