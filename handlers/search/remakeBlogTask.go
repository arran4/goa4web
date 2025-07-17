package search

import (
	"fmt"
	"net/http"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeBlogTask rebuilds the blog search index.
type RemakeBlogTask struct{ tasks.TaskString }

var remakeBlogTask = &RemakeBlogTask{TaskString: TaskRemakeBlogSearch}

func (RemakeBlogTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	data := struct {
		*handlers.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData),
		Back:     "/admin/search",
	}
	if err := queries.DeleteBlogsSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteBlogsSearch: %w", err).Error())
	}
	if err := queries.RemakeBlogsSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeBlogsSearchInsert: %w", err).Error())
	}

	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
