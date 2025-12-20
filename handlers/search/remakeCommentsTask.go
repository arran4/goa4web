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

// RemakeCommentsTask rebuilds the comments search index.
type RemakeCommentsTask struct{ tasks.TaskString }

var remakeCommentsTask = &RemakeCommentsTask{TaskString: TaskRemakeCommentsSearch}
var _ tasks.Task = (*RemakeCommentsTask)(nil)
var _ tasks.BackgroundTasker = (*RemakeCommentsTask)(nil)
var _ tasks.TemplatesRequired = (*RemakeCommentsTask)(nil)

func (RemakeCommentsTask) Action(w http.ResponseWriter, r *http.Request) any {
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

func (RemakeCommentsTask) TemplatesRequired() []string {
	return []string{handlers.TemplateRunTaskPage}
}

func (RemakeCommentsTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
	if err := q.SystemDeleteCommentsSearch(ctx); err != nil {
		return nil, err
	}
	rows, err := q.GetAllCommentsForIndex(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllCommentsForIndex: %w", err)
	}
	cache := map[string]int64{}
	for _, row := range rows {
		text := strings.TrimSpace(row.Text.String)
		if text == "" {
			continue
		}
		if err := indexText(ctx, q, cache, text, func(c context.Context, wid int64, count int32) error {
			return q.SystemAddToForumCommentSearch(c, db.SystemAddToForumCommentSearchParams{
				CommentID:                      row.Idcomments,
				SearchwordlistIdsearchwordlist: int32(wid),
				WordCount:                      count,
			})
		}); err != nil {
			return nil, err
		}
		if err := q.SystemSetCommentLastIndex(ctx, row.Idcomments); err != nil {
			return nil, err
		}
	}
	return remakeCommentsFinishedTask, nil
}
