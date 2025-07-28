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

// RemakeNewsTask rebuilds the news search index.
type RemakeNewsTask struct{ tasks.TaskString }

var remakeNewsTask = &RemakeNewsTask{TaskString: TaskRemakeNewsSearch}
var _ tasks.Task = (*RemakeNewsTask)(nil)
var _ tasks.BackgroundTasker = (*RemakeNewsTask)(nil)

func (RemakeNewsTask) Action(w http.ResponseWriter, r *http.Request) any {
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

func (RemakeNewsTask) BackgroundTask(ctx context.Context, q *dbpkg.Queries) (tasks.Task, error) {
	if err := q.DeleteSiteNewsSearch(ctx); err != nil {
		return nil, err
	}
	rows, err := q.GetAllSiteNewsForIndex(ctx)
	if err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetAllSiteNewsForIndex: %w", err).Error())
	} else {
		cache := map[string]int64{}
		for _, row := range rows {
			text := strings.TrimSpace(row.News.String)
			if text == "" {
				continue
			}
			err := indexText(ctx, q, cache, text, func(c context.Context, wid int64, count int32) error {
				return queries.AddToSiteNewsSearch(c, db.AddToSiteNewsSearchParams{
					SiteNewsID:                     row.Idsitenews,
					SearchwordlistIdsearchwordlist: int32(wid),
					WordCount:                      count,
				})
			})
		}); err != nil {
			return nil, err
		}
		if err := q.SetSiteNewsLastIndex(ctx, row.Idsitenews); err != nil {
			return nil, err
		}
	}
	return remakeNewsFinishedTask, nil
}
