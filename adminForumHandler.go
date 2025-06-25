package goa4web

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver.
)

func adminForumPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
	}
	err := templates.RenderTemplate(w, "page.gohtml", data, common.NewFuncs(r))
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminForumRemakeForumThreadPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
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
	err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, common.NewFuncs(r))
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminForumRemakeForumTopicPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
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
	err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, common.NewFuncs(r))
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func countForumThreads(ctx context.Context, q *Queries) (int64, error) {
	var c int64
	err := q.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM forumthread").Scan(&c)
	return c, err
}

func countForumTopics(ctx context.Context, q *Queries) (int64, error) {
	var c int64
	err := q.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM forumtopic").Scan(&c)
	return c, err
}
