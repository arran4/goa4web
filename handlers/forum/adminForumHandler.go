package forum

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

)

func AdminForumPage(w http.ResponseWriter, r *http.Request) {
	type Stats struct {
		Categories int64
		Topics     int64
		Threads    int64
	}

	type Data struct {
		Categories              []*ForumcategoryPlus
		CategoryBreadcrumbs     []*ForumcategoryPlus
		Admin                   bool
		CopyDataToSubCategories func(rootCategory *ForumcategoryPlus) *Data
		Category                *ForumcategoryPlus
		Back                    bool
		Stats                   Stats
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum Admin"
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
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}


	queries := cd.Queries()

	rows, err := cd.ForumTopics(0)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("ForumTopics Error: %s", err)
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	var topicRows []*ForumtopicPlus
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
			Lastposterusername:           row.Lastposterusername,
			Labels:                       nil,
		})
	}

	categoryTree := NewCategoryTree(categoryRows, topicRows)
	data.Categories = categoryTree.CategoryChildrenLookup[0]

	stats, err := queries.AdminGetForumStats(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("database not available"))
		return
	}
	data.Stats.Categories = stats.Categories
	data.Stats.Topics = stats.Topics
	data.Stats.Threads = stats.Threads

	ForumAdminPageTmpl.Handle(w, r, data)
}

const ForumAdminPageTmpl tasks.Template = "forum/adminPage.gohtml"

func AdminForumRemakeForumThreadPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum - Rebuild Threads"
	queries := cd.Queries()
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: "/admin/forum",
	}

	if c, err := countForumThreads(r.Context(), queries); err == nil {
		data.Messages = append(data.Messages, fmt.Sprintf("Processing %d threads...", c))
	}
	data.Messages = append(data.Messages, "Recalculating forum thread metadata...")

	if err := queries.AdminRecalculateAllForumThreadMetaData(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("recalculateForumThreadByIdMetaData_firstpost: %w", err).Error())
	} else {
		data.Messages = append(data.Messages, "Thread metadata rebuild complete.")
	}
	RunTaskPageTmpl.Handle(w, r, data)
}

func AdminForumRemakeForumTopicPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum - Rebuild Topics"
	queries := cd.Queries()
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: "/admin/forum",
	}

	if c, err := countForumTopics(r.Context(), queries); err == nil {
		data.Messages = append(data.Messages, fmt.Sprintf("Processing %d topics...", c))
	}

	if err := queries.AdminRebuildAllForumTopicMetaColumns(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("rebuildForumTopicByIdMetaColumns_lastaddition_lastposter: %w", err).Error())
	} else {
		data.Messages = append(data.Messages, "Topic metadata rebuild complete.")
	}
	RunTaskPageTmpl.Handle(w, r, data)
}

const RunTaskPageTmpl tasks.Template = "admin/runTaskPage.gohtml"

func countForumThreads(ctx context.Context, q db.Querier) (int64, error) {
	return q.AdminCountForumThreads(ctx)
}

func countForumTopics(ctx context.Context, q db.Querier) (int64, error) {
	return q.AdminCountForumTopics(ctx)
}
