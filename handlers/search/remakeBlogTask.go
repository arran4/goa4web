package search

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeBlogTask rebuilds the blog search index.
type RemakeBlogTask struct{ tasks.TaskString }

var remakeBlogTask = &RemakeBlogTask{TaskString: TaskRemakeBlogSearch}
var _ tasks.Task = (*RemakeBlogTask)(nil)
var _ tasks.BackgroundTasker = (*RemakeBlogTask)(nil)

func (RemakeBlogTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: cd,
		Messages: []string{"work queued"},
		Back:     "/admin/search",
	}
	return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
}

func (RemakeBlogTask) BackgroundTask(ctx context.Context, q *dbpkg.Queries) (tasks.Task, error) {
	if err := q.DeleteBlogsSearch(ctx); err != nil {
		return nil, err
	}
	rows, err := q.SystemGetAllBlogsForIndex(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllBlogsForIndex: %w", err)
	}
	cache := map[string]int64{}
	for _, row := range rows {
		text := strings.TrimSpace(row.Blog.String)
		if text == "" {
			continue
		}
		if err := indexText(ctx, q, cache, text, func(c context.Context, wid int64, count int32) error {
			return q.AddToBlogsSearch(c, dbpkg.AddToBlogsSearchParams{
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
