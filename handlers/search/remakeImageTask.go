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

// RemakeImageTask rebuilds the image search index.
type RemakeImageTask struct{ tasks.TaskString }

var remakeImageTask = &RemakeImageTask{TaskString: TaskRemakeImageSearch}
var _ tasks.Task = (*RemakeImageTask)(nil)
var _ tasks.BackgroundTasker = (*RemakeImageTask)(nil)
var _ tasks.TemplatesRequired = (*RemakeImageTask)(nil)

func (RemakeImageTask) Action(w http.ResponseWriter, r *http.Request) any {
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

func (RemakeImageTask) RequiredTemplates() []tasks.Template {
	return []tasks.Template{tasks.Template(handlers.TemplateRunTaskPage)}
}

func (RemakeImageTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
	if err := q.SystemDeleteImagePostSearch(ctx); err != nil {
		return nil, err
	}
	rows, err := q.GetAllImagePostsForIndex(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllImagePostsForIndex: %w", err)
	}
	cache := map[string]int64{}
	for _, row := range rows {
		text := strings.TrimSpace(row.Description.String)
		if text == "" {
			continue
		}
		if err := indexText(ctx, q, cache, text, func(c context.Context, wid int64, count int32) error {
			return q.SystemAddToImagePostSearch(c, db.SystemAddToImagePostSearchParams{
				ImagePostID:                    row.Idimagepost,
				SearchwordlistIdsearchwordlist: int32(wid),
				WordCount:                      count,
			})
		}); err != nil {
			return nil, err
		}
		if err := q.SystemSetImagePostLastIndex(ctx, row.Idimagepost); err != nil {
			return nil, err
		}
	}
	return remakeImageFinishedTask, nil
}
