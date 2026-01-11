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

// RemakeLinkerTask rebuilds the linker search index.
type RemakeLinkerTask struct{ tasks.TaskString }

var remakeLinkerTask = &RemakeLinkerTask{TaskString: TaskRemakeLinkerSearch}
var _ tasks.Task = (*RemakeLinkerTask)(nil)
var _ tasks.BackgroundTasker = (*RemakeLinkerTask)(nil)
var _ tasks.TemplatesRequired = (*RemakeLinkerTask)(nil)

func (RemakeLinkerTask) Action(w http.ResponseWriter, r *http.Request) any {
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

func (RemakeLinkerTask) TemplatesRequired() []tasks.Page {
	return []tasks.Page{handlers.TemplateRunTaskPage}
}

func (RemakeLinkerTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
	if err := q.SystemDeleteLinkerSearch(ctx); err != nil {
		return nil, err
	}
	rows, err := q.GetAllLinkersForIndex(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllLinkersForIndex: %w", err)
	}
	cache := map[string]int64{}
	for _, row := range rows {
		text := strings.TrimSpace(row.Title.String + " " + row.Description.String)
		if text == "" {
			continue
		}
		if err := indexText(ctx, q, cache, text, func(c context.Context, wid int64, count int32) error {
			return q.SystemAddToLinkerSearch(c, db.SystemAddToLinkerSearchParams{
				LinkerID:                       row.ID,
				SearchwordlistIdsearchwordlist: int32(wid),
				WordCount:                      count,
			})
		}); err != nil {
			return nil, err
		}
		if err := q.SystemSetLinkerLastIndex(ctx, row.ID); err != nil {
			return nil, err
		}
	}
	return remakeLinkerFinishedTask, nil
}
