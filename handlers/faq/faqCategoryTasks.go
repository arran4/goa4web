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

// DeleteCategoryTask removes a category.
type DeleteCategoryTask struct{ tasks.TaskString }

// CreateCategoryTask adds a new category entry.
type CreateCategoryTask struct{ tasks.TaskString }

var renameCategoryTask = &RenameCategoryTask{TaskString: TaskRenameCategory}
var _ tasks.Task = (*RenameCategoryTask)(nil)
var deleteCategoryTask = &DeleteCategoryTask{TaskString: TaskDeleteCategory}
var _ tasks.Task = (*DeleteCategoryTask)(nil)
var createCategoryTask = &CreateCategoryTask{TaskString: TaskCreateCategory}
var _ tasks.Task = (*CreateCategoryTask)(nil)

func (RenameCategoryTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskRenameCategory)(r, m)
}

func (DeleteCategoryTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskDeleteCategory)(r, m)
}

func (CreateCategoryTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskCreateCategory)(r, m)
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

func (DeleteCategoryTask) Action(w http.ResponseWriter, r *http.Request) any {
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		return fmt.Errorf("cid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if err := queries.DeleteFAQCategory(r.Context(), int32(cid)); err != nil {
		return fmt.Errorf("delete category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}

func (CreateCategoryTask) Action(w http.ResponseWriter, r *http.Request) any {
	text := r.PostFormValue("cname")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if err := queries.CreateFAQCategory(r.Context(), sql.NullString{
		String: text,
		Valid:  true,
	}); err != nil {
		return fmt.Errorf("create category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}
