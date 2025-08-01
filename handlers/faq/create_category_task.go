package faq

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// CreateCategoryTask adds a new category entry.
type CreateCategoryTask struct{ tasks.TaskString }

var createCategoryTask = &CreateCategoryTask{TaskString: TaskCreateCategory}
var _ tasks.Task = (*CreateCategoryTask)(nil)

func (CreateCategoryTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskCreateCategory)(r, m)
}

func (CreateCategoryTask) Action(w http.ResponseWriter, r *http.Request) any {
	text := r.PostFormValue("cname")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if err := queries.AdminCreateFAQCategory(r.Context(), sql.NullString{
		String: text,
		Valid:  true,
	}); err != nil {
		return fmt.Errorf("create category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
