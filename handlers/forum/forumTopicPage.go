package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/core"
	"github.com/gorilla/mux"
)

func TopicsPageWithBasePath(w http.ResponseWriter, r *http.Request, basePath string) {
	type Data struct {
		Admin                   bool
		Back                    bool
		Subscribed              bool
		Topic                   *ForumtopicPlus
		Threads                 []*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow
		Categories              []*ForumcategoryPlus
		Category                *ForumcategoryPlus
		CopyDataToSubCategories func(rootCategory *ForumcategoryPlus) *Data
		BasePath                string
		PublicLabels            []string
		AuthorLabels            []string
		PrivateLabels           []string
	}

	if _, ok := core.GetSessionOrFail(w, r); !ok {
		return
	}
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	cd.ForumBasePath = basePath
	data := &Data{
		Admin:    cd.IsAdmin() && cd.IsAdminMode(),
		BasePath: basePath,
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
	displayTitle := topicRow.Title.String
	if topicRow.Handler == "private" && cd.Queries() != nil {
		parts, err := cd.Queries().ListPrivateTopicParticipantsByTopicIDForUser(r.Context(), db.ListPrivateTopicParticipantsByTopicIDForUserParams{
			TopicID:  sql.NullInt32{Int32: topicRow.Idforumtopic, Valid: true},
			ViewerID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
		if err != nil {
			log.Printf("list private participants: %v", err)
		}
		var names []string
		for _, p := range parts {
			if p.Idusers != cd.UserID {
				names = append(names, p.Username.String)
			}
		}
		if len(names) > 0 {
			displayTitle = strings.Join(names, ", ")
		}
	}
	cd.PageTitle = fmt.Sprintf("Forum - %s", displayTitle)
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
		DisplayTitle:                 displayTitle,
		Edit:                         false,
	}

	if topicRow.Handler != "private" {
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
	}

	threadRows, err := cd.ForumThreads(int32(topicId))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Error: ForumThreads: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	data.Threads = threadRows

	if pub, author, err := cd.TopicPublicLabels(topicRow.Idforumtopic); err == nil {
		data.PublicLabels = pub
		data.AuthorLabels = author
	} else {
		log.Printf("list public labels: %v", err)
	}
	if priv, err := cd.TopicPrivateLabels(topicRow.Idforumtopic); err == nil {
		data.PrivateLabels = priv
	} else {
		log.Printf("list private labels: %v", err)
	}

	if subscribedToTopic(cd, topicRow.Idforumtopic) {
		data.Subscribed = true
	}

	handlers.TemplateHandler(w, r, "topicsPage.gohtml", data)
}

// TopicsPage serves the forum topic page at the default /forum prefix.
func TopicsPage(w http.ResponseWriter, r *http.Request) {
	TopicsPageWithBasePath(w, r, "/forum")
}
