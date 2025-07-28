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

// RemakeBlogTask rebuilds the blog search index.
type RemakeBlogTask struct{ tasks.TaskString }

var remakeBlogTask = &RemakeBlogTask{TaskString: TaskRemakeBlogSearch}
var _ tasks.Task = (*RemakeBlogTask)(nil)

func (RemakeBlogTask) Action(w http.ResponseWriter, r *http.Request) any {
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
	if err := queries.DeleteBlogsSearch(ctx); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteBlogsSearch: %w", err).Error())
	}

	rows, err := queries.GetAllBlogsForIndex(ctx)
	if err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetAllBlogsForIndex: %w", err).Error())
	} else {
		cache := map[string]int64{}
		for _, row := range rows {
			text := strings.TrimSpace(row.Blog.String)
			if text == "" {
				continue
			}
			err := indexText(ctx, queries, cache, text, func(c context.Context, wid int64, count int32) error {
				return queries.AddToBlogsSearch(c, db.AddToBlogsSearchParams{
					BlogID:                         row.Idblogs,
					SearchwordlistIdsearchwordlist: int32(wid),
					WordCount:                      count,
				})
			})
			if err != nil {
				data.Errors = append(data.Errors, fmt.Errorf("index blog %d: %w", row.Idblogs, err).Error())
				continue
			}
			if err := queries.SetBlogLastIndex(ctx, row.Idblogs); err != nil {
				data.Errors = append(data.Errors, fmt.Errorf("SetBlogLastIndex %d: %w", row.Idblogs, err).Error())
			}
		}
	}

	return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
}
