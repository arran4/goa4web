package search

import (
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	common "github.com/arran4/goa4web/core/common"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeNewsTask rebuilds the news search index.
type RemakeNewsTask struct{ tasks.TaskString }

var remakeNewsTask = &RemakeNewsTask{TaskString: TaskRemakeNewsSearch}
var _ tasks.Task = (*RemakeNewsTask)(nil)

func (RemakeNewsTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/search",
	}
	if err := queries.DeleteSiteNewsSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteSiteNewsSearch: %w", err).Error())
	}
	if err := queries.RemakeNewsSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeNewsSearchInsert: %w", err).Error())
	}

	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
