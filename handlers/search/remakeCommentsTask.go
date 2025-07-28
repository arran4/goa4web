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

// RemakeCommentsTask rebuilds the comments search index.
type RemakeCommentsTask struct{ tasks.TaskString }

var remakeCommentsTask = &RemakeCommentsTask{TaskString: TaskRemakeCommentsSearch}
var _ tasks.Task = (*RemakeCommentsTask)(nil)
var _ tasks.BackgroundTasker = (*RemakeCommentsTask)(nil)

func (RemakeCommentsTask) Action(w http.ResponseWriter, r *http.Request) any {
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

func (RemakeCommentsTask) BackgroundTask(ctx context.Context, q *dbpkg.Queries) (tasks.Task, error) {
	if err := q.DeleteCommentsSearch(ctx); err != nil {
		return nil, err
	}
	rows, err := q.GetAllCommentsForIndex(ctx)
	if err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetAllCommentsForIndex: %w", err).Error())
	} else {
		cache := map[string]int64{}
		for _, row := range rows {
			text := strings.TrimSpace(row.Text.String)
			if text == "" {
				continue
			}
			err := indexText(ctx, q, cache, text, func(c context.Context, wid int64, count int32) error {
				return queries.AddToForumCommentSearch(c, db.AddToForumCommentSearchParams{
					CommentID:                      row.Idcomments,
					SearchwordlistIdsearchwordlist: int32(wid),
					WordCount:                      count,
				})
			})
		}); err != nil {
			return nil, err
		}
		if err := q.SetCommentLastIndex(ctx, row.Idcomments); err != nil {
			return nil, err
		}
	}
	return remakeCommentsFinishedTask, nil
}
