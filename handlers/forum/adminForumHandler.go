package forum

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminForumPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	handlers.TemplateHandler(w, r, "forumAdminPage", data)
}

func AdminForumRemakeForumThreadPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
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
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
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
