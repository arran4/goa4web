package search

import (
	"fmt"
	"net/http"

	common "github.com/arran4/goa4web/core/common"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeCommentsTask rebuilds the comments search index.
type RemakeCommentsTask struct{ tasks.TaskString }

var remakeCommentsTask = &RemakeCommentsTask{TaskString: TaskRemakeCommentsSearch}

func (RemakeCommentsTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
		Back:     "/admin/search",
	}
	if err := queries.DeleteCommentsSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteCommentsSearch: %w", err).Error())
	}
	if err := queries.RemakeCommentsSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeCommentsSearchInsert: %w", err).Error())
	}

	handlers.TemplateHandler(w, r, "tasks/run_task.gohtml", data)
}
