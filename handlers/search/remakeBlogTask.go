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

// RemakeBlogTask rebuilds the blog search index.
type RemakeBlogTask struct{ tasks.TaskString }

var remakeBlogTask = &RemakeBlogTask{TaskString: TaskRemakeBlogSearch}
var _ tasks.Task = (*RemakeBlogTask)(nil)
var _ tasks.BackgroundTasker = (*RemakeBlogTask)(nil)
var _ tasks.TemplatesRequired = (*RemakeBlogTask)(nil)

func (RemakeBlogTask) Action(w http.ResponseWriter, r *http.Request) any {
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

func (RemakeBlogTask) TemplatesRequired() []string {
	return []string{handlers.TemplateRunTaskPage}
}

func (RemakeBlogTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
	if err := q.SystemDeleteBlogsSearch(ctx); err != nil {
		return nil, err
	}
	rows, err := q.SystemGetAllBlogsForIndex(ctx)
	if err != nil {
		return nil, fmt.Errorf("SystemGetAllBlogsForIndex: %w", err)
	}
	cache := map[string]int64{}
	for _, row := range rows {
		text := strings.TrimSpace(row.Blog.String)
		if text == "" {
			continue
		}
		if err := indexText(ctx, q, cache, text, func(c context.Context, wid int64, count int32) error {
			return q.SystemAddToBlogsSearch(c, db.SystemAddToBlogsSearchParams{
				BlogID:                         row.Idblogs,
				SearchwordlistIdsearchwordlist: int32(wid),
				WordCount:                      count,
			})
		}); err != nil {
			return nil, err
		}
		if err := q.SystemSetBlogLastIndex(ctx, row.Idblogs); err != nil {
			return nil, err
		}
	}
	return remakeBlogFinishedTask, nil
}
