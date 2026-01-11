package search

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeNewsTask rebuilds the news search index.
type RemakeNewsTask struct{ tasks.TaskString }

var remakeNewsTask = &RemakeNewsTask{TaskString: TaskRemakeNewsSearch}
var _ tasks.Task = (*RemakeNewsTask)(nil)
var _ tasks.BackgroundTasker = (*RemakeNewsTask)(nil)
var _ tasks.TemplatesRequired = (*RemakeNewsTask)(nil)

func (RemakeNewsTask) Action(w http.ResponseWriter, r *http.Request) any {
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Messages: []string{"work queued"},
		Back:     "/admin/search",
	}
	return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
}

func (RemakeNewsTask) TemplatesRequired() []tasks.Page {
	return []tasks.Page{handlers.TemplateRunTaskPage}
}

func (RemakeNewsTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
	if err := q.SystemDeleteSiteNewsSearch(ctx); err != nil {
		return nil, err
	}
	rows, err := q.GetAllSiteNewsForIndex(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllSiteNewsForIndex: %w", err)
	}
	cache := map[string]int64{}
	for _, row := range rows {
		text := strings.TrimSpace(row.News.String)
		if text == "" {
			continue
		}
		if err := indexText(ctx, q, cache, text, func(c context.Context, wid int64, count int32) error {
			return q.SystemAddToSiteNewsSearch(c, db.SystemAddToSiteNewsSearchParams{
				SiteNewsID:                     row.Idsitenews,
				SearchwordlistIdsearchwordlist: int32(wid),
				WordCount:                      count,
			})
		}); err != nil {
			return nil, err
		}
		if err := q.SystemSetSiteNewsLastIndex(ctx, row.Idsitenews); err != nil {
			return nil, err
		}
	}
	return remakeNewsFinishedTask, nil
}
