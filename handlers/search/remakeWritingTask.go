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

// RemakeWritingTask rebuilds the writing search index.
type RemakeWritingTask struct{ tasks.TaskString }

var remakeWritingTask = &RemakeWritingTask{TaskString: TaskRemakeWritingSearch}
var _ tasks.Task = (*RemakeWritingTask)(nil)
var _ tasks.BackgroundTasker = (*RemakeWritingTask)(nil)
var _ tasks.TemplatesRequired = (*RemakeWritingTask)(nil)

func (RemakeWritingTask) Action(w http.ResponseWriter, r *http.Request) any {
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

func (RemakeWritingTask) TemplatesRequired() []tasks.Page {
	return []tasks.Page{handlers.TemplateRunTaskPage}
}

func (RemakeWritingTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
	if err := q.SystemDeleteWritingSearch(ctx); err != nil {
		return nil, err
	}
	rows, err := q.GetAllWritingsForIndex(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllWritingsForIndex: %w", err)
	}
	cache := map[string]int64{}
	for _, row := range rows {
		text := strings.TrimSpace(row.Title.String + " " + row.Abstract.String + " " + row.Writing.String)
		if text == "" {
			continue
		}
		if err := indexText(ctx, q, cache, text, func(c context.Context, wid int64, count int32) error {
			return q.SystemAddToForumWritingSearch(c, db.SystemAddToForumWritingSearchParams{
				WritingID:                      row.Idwriting,
				SearchwordlistIdsearchwordlist: int32(wid),
				WordCount:                      count,
			})
		}); err != nil {
			return nil, err
		}
		if err := q.SystemSetWritingLastIndex(ctx, row.Idwriting); err != nil {
			return nil, err
		}
	}
	return remakeWritingFinishedTask, nil
}
