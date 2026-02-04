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

// UpdateCategoryTask updates a category.
type UpdateCategoryTask struct{ tasks.TaskString }

var updateCategoryTask = &UpdateCategoryTask{TaskString: TaskRenameCategory} // Keeping TaskRenameCategory string for now or should I change it?
// The route matcher uses TaskRenameCategory. If I change variable name, I should check handlers/faq/routes.go.
// I will update TaskString constant if I can find where it is defined. It is likely in `tasks` package or `faq` package (unexported).
// `TaskRenameCategory` is likely a constant in this package (I saw `renameCategoryTask` using `TaskRenameCategory`).
// Let's assume it is defined in `handlers/faq/tasks.go` or similar (if exists) or just unexported in this package.
// I'll stick to `TaskRenameCategory` for the task string for compatibility unless I change `routes.go` matcher too.
// Wait, `handlers/faq/routes.go` matches `renameCategoryTask.Matcher()`.
// I will rename the variable `renameCategoryTask` to `updateCategoryTask` in `update_category_task.go` and update `routes.go`.

var _ tasks.Task = (*UpdateCategoryTask)(nil)

func (UpdateCategoryTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskRenameCategory)(r, m)
}

func (UpdateCategoryTask) Action(w http.ResponseWriter, r *http.Request) any {
	text := r.PostFormValue("cname")
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		return fmt.Errorf("cid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	parentID, _ := strconv.Atoi(r.PostFormValue("parent_id"))
	priority, _ := strconv.Atoi(r.PostFormValue("priority"))

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if err := cd.UpdateFAQCategory(int32(cid), text, int32(parentID), int32(priority)); err != nil {
		return fmt.Errorf("update faq category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
