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

// RemakeLinkerTask rebuilds the linker search index.
type RemakeLinkerTask struct{ tasks.TaskString }

var remakeLinkerTask = &RemakeLinkerTask{TaskString: TaskRemakeLinkerSearch}
var _ tasks.Task = (*RemakeLinkerTask)(nil)

func (RemakeLinkerTask) Action(w http.ResponseWriter, r *http.Request) any {
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
	if err := queries.DeleteLinkerSearch(ctx); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteLinkerSearch: %w", err).Error())
	}

	rows, err := queries.GetAllLinkersForIndex(ctx)
	if err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetAllLinkersForIndex: %w", err).Error())
	} else {
		cache := map[string]int64{}
		for _, row := range rows {
			text := strings.TrimSpace(row.Title.String + " " + row.Description.String)
			if text == "" {
				continue
			}
			err := indexText(ctx, queries, cache, text, func(c context.Context, wid int64, count int32) error {
				return queries.AddToLinkerSearch(c, db.AddToLinkerSearchParams{
					LinkerID:                       row.Idlinker,
					SearchwordlistIdsearchwordlist: int32(wid),
					WordCount:                      count,
				})
			})
			if err != nil {
				data.Errors = append(data.Errors, fmt.Errorf("index linker %d: %w", row.Idlinker, err).Error())
				continue
			}
			if err := queries.SetLinkerLastIndex(ctx, row.Idlinker); err != nil {
				data.Errors = append(data.Errors, fmt.Errorf("SetLinkerLastIndex %d: %w", row.Idlinker, err).Error())
			}
		}
	}

	return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
}
