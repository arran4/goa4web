package search

import (
	"context"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeWritingTask rebuilds the writing search index.
type RemakeWritingTask struct{ tasks.TaskString }

var remakeWritingTask = &RemakeWritingTask{TaskString: TaskRemakeWritingSearch}
var _ tasks.Task = (*RemakeWritingTask)(nil)
var _ tasks.BackgroundTasker = (*RemakeWritingTask)(nil)

func (RemakeWritingTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := struct {
		*common.CoreData
		Messages []string
		Back     string
	}{
		CoreData: cd,
		Messages: []string{"work queued"},
		Back:     "/admin/search",
	}
	return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
}

func (RemakeWritingTask) BackgroundTask(ctx context.Context, q *dbpkg.Queries) (tasks.Task, error) {
	if err := q.DeleteWritingSearch(ctx); err != nil {
		return nil, err
	}
	rows, err := q.GetAllWritingsForIndex(ctx)
	if err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetAllWritingsForIndex: %w", err).Error())
	} else {
		cache := map[string]int64{}
		for _, row := range rows {
			text := strings.TrimSpace(row.Title.String + " " + row.Abstract.String + " " + row.Writing.String)
			if text == "" {
				continue
			}
			err := indexText(ctx, q, cache, text, func(c context.Context, wid int64, count int32) error {
				return queries.AddToForumWritingSearch(c, db.AddToForumWritingSearchParams{
					WritingID:                      row.Idwriting,
					SearchwordlistIdsearchwordlist: int32(wid),
					WordCount:                      count,
				})
			})
		}); err != nil {
			return nil, err
		}
		if err := q.SetWritingLastIndex(ctx, row.Idwriting); err != nil {
			return nil, err
		}
	}
	return remakeWritingFinishedTask, nil
}
