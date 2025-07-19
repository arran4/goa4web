package search

import (
	"fmt"
	"net/http"

	common "github.com/arran4/goa4web/core/common"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeLinkerTask rebuilds the linker search index.
type RemakeLinkerTask struct{ tasks.TaskString }

var remakeLinkerTask = &RemakeLinkerTask{TaskString: TaskRemakeLinkerSearch}

func (RemakeLinkerTask) Action(w http.ResponseWriter, r *http.Request) {
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
	if err := queries.DeleteLinkerSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteLinkerSearch: %w", err).Error())
	}
	if err := queries.RemakeLinkerSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeLinkerSearchInsert: %w", err).Error())
	}

	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
