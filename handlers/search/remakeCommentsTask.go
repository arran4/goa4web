package search

import (
	"context"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeCommentsTask rebuilds the comments search index.
type RemakeCommentsTask struct{ tasks.TaskString }

var remakeCommentsTask = &RemakeCommentsTask{TaskString: TaskRemakeCommentsSearch}
var _ tasks.Task = (*RemakeCommentsTask)(nil)

func (RemakeCommentsTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/search",
	}
	ctx := r.Context()
	if err := queries.DeleteCommentsSearch(ctx); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteCommentsSearch: %w", err).Error())
	}

	rows, err := queries.GetAllCommentsForIndex(ctx)
	if err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetAllCommentsForIndex: %w", err).Error())
	} else {
		cache := map[string]int64{}
		for _, row := range rows {
			text := strings.TrimSpace(row.Text.String)
			if text == "" {
				continue
			}
			err := indexText(ctx, queries, cache, text, func(c context.Context, wid int64, count int32) error {
				return queries.AddToForumCommentSearch(c, db.AddToForumCommentSearchParams{
					CommentID:                      row.Idcomments,
					SearchwordlistIdsearchwordlist: int32(wid),
					WordCount:                      count,
				})
			})
			if err != nil {
				data.Errors = append(data.Errors, fmt.Errorf("index comment %d: %w", row.Idcomments, err).Error())
				continue
			}
			if err := queries.SetCommentLastIndex(ctx, row.Idcomments); err != nil {
				data.Errors = append(data.Errors, fmt.Errorf("SetCommentLastIndex %d: %w", row.Idcomments, err).Error())
			}
		}
	}

	return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
}
