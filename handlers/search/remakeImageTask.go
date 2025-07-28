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

// RemakeImageTask rebuilds the image search index.
type RemakeImageTask struct{ tasks.TaskString }

var remakeImageTask = &RemakeImageTask{TaskString: TaskRemakeImageSearch}
var _ tasks.Task = (*RemakeImageTask)(nil)

func (RemakeImageTask) Action(w http.ResponseWriter, r *http.Request) any {
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
	if err := queries.DeleteImagePostSearch(ctx); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteImagePostSearch: %w", err).Error())
	}

	rows, err := queries.GetAllImagePostsForIndex(ctx)
	if err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetAllImagePostsForIndex: %w", err).Error())
	} else {
		cache := map[string]int64{}
		for _, row := range rows {
			text := strings.TrimSpace(row.Description.String)
			if text == "" {
				continue
			}
			err := indexText(ctx, queries, cache, text, func(c context.Context, wid int64, count int32) error {
				return queries.AddToImagePostSearch(c, db.AddToImagePostSearchParams{
					ImagePostID:                    row.Idimagepost,
					SearchwordlistIdsearchwordlist: int32(wid),
					WordCount:                      count,
				})
			})
			if err != nil {
				data.Errors = append(data.Errors, fmt.Errorf("index image %d: %w", row.Idimagepost, err).Error())
				continue
			}
			if err := queries.SetImagePostLastIndex(ctx, row.Idimagepost); err != nil {
				data.Errors = append(data.Errors, fmt.Errorf("SetImagePostLastIndex %d: %w", row.Idimagepost, err).Error())
			}
		}
	}

	return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
}
