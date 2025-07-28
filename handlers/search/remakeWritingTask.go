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

// RemakeWritingTask rebuilds the writing search index.
type RemakeWritingTask struct{ tasks.TaskString }

var remakeWritingTask = &RemakeWritingTask{TaskString: TaskRemakeWritingSearch}
var _ tasks.Task = (*RemakeWritingTask)(nil)

func (RemakeWritingTask) Action(w http.ResponseWriter, r *http.Request) any {
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
	if err := queries.DeleteWritingSearch(ctx); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteWritingSearch: %w", err).Error())
	}

	rows, err := queries.GetAllWritingsForIndex(ctx)
	if err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetAllWritingsForIndex: %w", err).Error())
	} else {
		cache := map[string]int64{}
		for _, row := range rows {
			text := strings.TrimSpace(row.Title.String + " " + row.Abstract.String + " " + row.Writing.String)
			if text == "" {
				continue
			}
			err := indexText(ctx, queries, cache, text, func(c context.Context, wid int64, count int32) error {
				return queries.AddToForumWritingSearch(c, db.AddToForumWritingSearchParams{
					WritingID:                      row.Idwriting,
					SearchwordlistIdsearchwordlist: int32(wid),
					WordCount:                      count,
				})
			})
			if err != nil {
				data.Errors = append(data.Errors, fmt.Errorf("index writing %d: %w", row.Idwriting, err).Error())
				continue
			}
			if err := queries.SetWritingLastIndex(ctx, row.Idwriting); err != nil {
				data.Errors = append(data.Errors, fmt.Errorf("SetWritingLastIndex %d: %w", row.Idwriting, err).Error())
			}
		}
	}

	return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
}
