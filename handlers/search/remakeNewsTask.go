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

// RemakeNewsTask rebuilds the news search index.
type RemakeNewsTask struct{ tasks.TaskString }

var remakeNewsTask = &RemakeNewsTask{TaskString: TaskRemakeNewsSearch}
var _ tasks.Task = (*RemakeNewsTask)(nil)

func (RemakeNewsTask) Action(w http.ResponseWriter, r *http.Request) any {
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
	if err := queries.DeleteSiteNewsSearch(ctx); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteSiteNewsSearch: %w", err).Error())
	}

	rows, err := queries.GetAllSiteNewsForIndex(ctx)
	if err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetAllSiteNewsForIndex: %w", err).Error())
	} else {
		for _, row := range rows {
			text := strings.TrimSpace(row.News.String)
			if text == "" {
				continue
			}
			err := indexText(ctx, queries, text, func(c context.Context, wid int64, count int32) error {
				return queries.AddToSiteNewsSearch(c, db.AddToSiteNewsSearchParams{
					SiteNewsID:                     row.Idsitenews,
					SearchwordlistIdsearchwordlist: int32(wid),
					WordCount:                      count,
				})
			})
			if err != nil {
				data.Errors = append(data.Errors, fmt.Errorf("index site news %d: %w", row.Idsitenews, err).Error())
				continue
			}
			if err := queries.SetSiteNewsLastIndex(ctx, row.Idsitenews); err != nil {
				data.Errors = append(data.Errors, fmt.Errorf("SetSiteNewsLastIndex %d: %w", row.Idsitenews, err).Error())
			}
		}
	}

	return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
}
