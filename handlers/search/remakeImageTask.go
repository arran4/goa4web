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

// RemakeImageTask rebuilds the image search index.
type RemakeImageTask struct{ tasks.TaskString }

var remakeImageTask = &RemakeImageTask{TaskString: TaskRemakeImageSearch}
var _ tasks.Task = (*RemakeImageTask)(nil)
var _ tasks.BackgroundTasker = (*RemakeImageTask)(nil)

func (RemakeImageTask) Action(w http.ResponseWriter, r *http.Request) any {
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

func (RemakeImageTask) BackgroundTask(ctx context.Context, q *dbpkg.Queries) (tasks.Task, error) {
	if err := q.DeleteImagePostSearch(ctx); err != nil {
		return nil, err
	}
	rows, err := q.GetAllImagePostsForIndex(ctx)
	if err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetAllImagePostsForIndex: %w", err).Error())
	} else {
		cache := map[string]int64{}
		for _, row := range rows {
			text := strings.TrimSpace(row.Description.String)
			if text == "" {
				continue
			}
			err := indexText(ctx, q, cache, text, func(c context.Context, wid int64, count int32) error {
				return queries.AddToImagePostSearch(c, db.AddToImagePostSearchParams{
					ImagePostID:                    row.Idimagepost,
					SearchwordlistIdsearchwordlist: int32(wid),
					WordCount:                      count,
				})
			})
		}); err != nil {
			return nil, err
		}
		if err := q.SetImagePostLastIndex(ctx, row.Idimagepost); err != nil {
			return nil, err
		}
	}
	return remakeImageFinishedTask, nil
}
