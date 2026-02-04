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

// DeleteCategoryTask removes a category.
type DeleteCategoryTask struct{ tasks.TaskString }

var deleteCategoryTask = &DeleteCategoryTask{TaskString: TaskDeleteCategory}
var _ tasks.Task = (*DeleteCategoryTask)(nil)

func (DeleteCategoryTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskDeleteCategory)(r, m)
}

func (DeleteCategoryTask) Action(w http.ResponseWriter, r *http.Request) any {
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		return fmt.Errorf("cid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if !cd.HasGrant("faq", "category", "delete", int32(cid)) {
		return fmt.Errorf("permission denied %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("permission denied")))
	}

	if err := cd.DeleteFAQCategory(int32(cid)); err != nil {
		return fmt.Errorf("delete category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
