package faq

import (
	"fmt"
	"net/http"
	"strconv"

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
	priority, _ := strconv.Atoi(r.PostFormValue("priority"))
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if !cd.HasGrant("faq", "category", "create", 0) {
		return fmt.Errorf("permission denied %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("permission denied")))
	}

	if err := cd.CreateFAQCategory(text, int32(priority)); err != nil {
		return fmt.Errorf("create category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
