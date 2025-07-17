package search

import (
	"fmt"
	corecommon "github.com/arran4/goa4web/core/common"
	"net/http"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeWritingTask rebuilds the writing search index.
type RemakeWritingTask struct{ tasks.TaskString }

var remakeWritingTask = &RemakeWritingTask{TaskString: TaskRemakeWritingSearch}

func (RemakeWritingTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	data := struct {
		*corecommon.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*corecommon.CoreData),
		Back:     "/admin/search",
	}
	if err := queries.DeleteWritingSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteWritingSearch: %w", err).Error())
	}
	if err := queries.RemakeWritingSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeWritingSearchInsert: %w", err).Error())
	}

	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
