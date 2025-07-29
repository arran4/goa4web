package forum

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/core"
)

func AdminForumPage(w http.ResponseWriter, r *http.Request) {
	type Stats struct {
		Categories int64
		Topics     int64
		Threads    int64
	}

	type Data struct {
		*common.CoreData
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
		CoreData: cd,
		Admin:    cd.CanEditAny(),
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

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	queries := cd.Queries()
	rows, err := queries.GetAllForumTopicsForUser(r.Context(), db.GetAllForumTopicsForUserParams{
		ViewerID:      uid,
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("showTableTopics Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
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
		})
	}

	categoryTree := NewCategoryTree(categoryRows, topicRows)
	data.Categories = categoryTree.CategoryChildrenLookup[0]

	ctx := r.Context()
	count := func(q string, dest *int64) {
		if err := queries.DB().QueryRowContext(ctx, q).Scan(dest); err != nil && err != sql.ErrNoRows {
			log.Printf("forumAdminPage count query error: %v", err)
		}
	}
	count("SELECT COUNT(*) FROM forumcategory", &data.Stats.Categories)
	count("SELECT COUNT(*) FROM forumtopic", &data.Stats.Topics)
	count("SELECT COUNT(*) FROM forumthread", &data.Stats.Threads)

	handlers.TemplateHandler(w, r, "forumAdminPage", data)
}

func AdminForumRemakeForumThreadPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum - Rebuild Threads"
	queries := cd.Queries()
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: cd,
		Back:     "/admin/forum",
	}

	if c, err := countForumThreads(r.Context(), queries); err == nil {
		data.Messages = append(data.Messages, fmt.Sprintf("Processing %d threads...", c))
	}
	data.Messages = append(data.Messages, "Recalculating forum thread metadata...")

	if err := queries.RecalculateAllForumThreadMetaData(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("recalculateForumThreadByIdMetaData_firstpost: %w", err).Error())
	} else {
		data.Messages = append(data.Messages, "Thread metadata rebuild complete.")
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func AdminForumRemakeForumTopicPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum - Rebuild Topics"
	queries := cd.Queries()
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: cd,
		Back:     "/admin/forum",
	}

	if c, err := countForumTopics(r.Context(), queries); err == nil {
		data.Messages = append(data.Messages, fmt.Sprintf("Processing %d topics...", c))
	}

	if err := queries.RebuildAllForumTopicMetaColumns(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("rebuildForumTopicByIdMetaColumns_lastaddition_lastposter: %w", err).Error())
	} else {
		data.Messages = append(data.Messages, "Topic metadata rebuild complete.")
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func countForumThreads(ctx context.Context, q *db.Queries) (int64, error) {
	var c int64
	err := q.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM forumthread").Scan(&c)
	return c, err
}

func countForumTopics(ctx context.Context, q *db.Queries) (int64, error) {
	var c int64
	err := q.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM forumtopic").Scan(&c)
	return c, err
}
