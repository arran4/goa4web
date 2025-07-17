package search

import (
	"fmt"
	corecommon "github.com/arran4/goa4web/core/common"
	"net/http"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeImageTask rebuilds the image search index.
type RemakeImageTask struct{ tasks.TaskString }

var remakeImageTask = &RemakeImageTask{TaskString: TaskRemakeImageSearch}

func (RemakeImageTask) Action(w http.ResponseWriter, r *http.Request) {
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
	if err := queries.DeleteImagePostSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteImagePostSearch: %w", err).Error())
	}
	if err := queries.RemakeImagePostSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeImagePostSearchInsert: %w", err).Error())
	}

	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
