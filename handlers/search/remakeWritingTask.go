package search

import (
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
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
	if err := queries.DeleteWritingSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteWritingSearch: %w", err).Error())
	}
	if err := queries.RemakeWritingSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeWritingSearchInsert: %w", err).Error())
	}

	return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
}
