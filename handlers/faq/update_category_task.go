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

var updateCategoryTask = &UpdateCategoryTask{TaskString: TaskRenameCategory}

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

	if !cd.HasGrant("faq", "category", "edit", int32(cid)) {
		return fmt.Errorf("permission denied %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("permission denied")))
	}

	if err := cd.UpdateFAQCategory(int32(cid), text, int32(parentID), int32(priority)); err != nil {
		return fmt.Errorf("update faq category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
