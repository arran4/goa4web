package faq

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// RenameCategoryTask renames a category.
type RenameCategoryTask struct{ tasks.TaskString }

var renameCategoryTask = &RenameCategoryTask{TaskString: TaskRenameCategory}
var _ tasks.Task = (*RenameCategoryTask)(nil)

func (RenameCategoryTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskRenameCategory)(r, m)
}

func (RenameCategoryTask) Action(w http.ResponseWriter, r *http.Request) any {
	text := r.PostFormValue("cname")
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		return fmt.Errorf("cid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if err := queries.RenameFAQCategory(r.Context(), db.RenameFAQCategoryParams{
		Name: sql.NullString{
			String: text,
			Valid:  true,
		},
		Idfaqcategories: int32(cid),
	}); err != nil {
		return fmt.Errorf("rename faq category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
