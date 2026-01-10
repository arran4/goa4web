package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/core"
	"github.com/gorilla/mux"
)

func TopicsPageWithBasePath(w http.ResponseWriter, r *http.Request, basePath string) {
	type threadWithLabels struct {
		*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow
		Labels   []templates.TopicLabel
		IsUnread bool
	}

	type Data struct {
		Admin                   bool
		Back                    bool
		Subscribed              bool
		Topic                   *ForumtopicPlus
		Threads                 []*threadWithLabels
		Categories              []*ForumcategoryPlus
		Category                *ForumcategoryPlus
		CopyDataToSubCategories func(rootCategory *ForumcategoryPlus) *Data
		BasePath                string
		BackURL                 string
		ShareURL                string
		Labels                  []templates.TopicLabel
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
		BackURL:  r.URL.RequestURI(),
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
			handlers.RenderErrorPage(w, r, handlers.ErrNotFound) // Use consistent error page helper
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
		Labels:                       nil,
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
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	threads := make([]*threadWithLabels, len(threadRows))
	for i, r := range threadRows {
		t := &threadWithLabels{GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow: r}
		var lbls []templates.TopicLabel
		if pub, author, err := cd.ThreadPublicLabels(r.Idforumthread); err == nil {
			for _, l := range pub {
				lbls = append(lbls, templates.TopicLabel{Name: l, Type: "public"})
			}
			for _, l := range author {
				lbls = append(lbls, templates.TopicLabel{Name: l, Type: "author"})
			}
		} else {
			log.Printf("list public labels: %v", err)
		}
		if priv, err := cd.ThreadPrivateLabels(r.Idforumthread, r.Firstpostuserid.Int32); err == nil {
			for _, l := range priv {
				lbls = append(lbls, templates.TopicLabel{Name: l, Type: "private"})
				if l == "unread" {
					t.IsUnread = true
				}
			}
		} else {
			log.Printf("list private labels: %v", err)
		}
		sort.Slice(lbls, func(i, j int) bool { return lbls[i].Name < lbls[j].Name })
		t.Labels = lbls
		threads[i] = t
	}
	data.Threads = threads

	var labels []templates.TopicLabel
	if pub, _, err := cd.ThreadPublicLabels(topicRow.Idforumtopic); err == nil {
		for _, l := range pub {
			labels = append(labels, templates.TopicLabel{Name: l, Type: "public"})
		}
	} else {
		log.Printf("list public labels: %v", err)
	}
	sort.Slice(labels, func(i, j int) bool { return labels[i].Name < labels[j].Name })
	data.Labels = labels

	if subscribedToTopic(cd, topicRow.Idforumtopic) {
		data.Subscribed = true
	}

	ForumTopicsPageTmpl.Handle(w, r, data)
}

const ForumTopicsPageTmpl handlers.Page = "forum/topicsPage.gohtml"

// TopicsPage serves the forum topic page at the default /forum prefix.
func TopicsPage(w http.ResponseWriter, r *http.Request) {
	TopicsPageWithBasePath(w, r, "/forum")
}
