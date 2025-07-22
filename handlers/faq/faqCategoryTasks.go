package faq

import (
	"database/sql"
	"log"
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

func (RenameCategoryTask) Action(w http.ResponseWriter, r *http.Request) {
	text := r.PostFormValue("cname")
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if err := queries.RenameFAQCategory(r.Context(), db.RenameFAQCategoryParams{
		Name: sql.NullString{
			String: text,
			Valid:  true,
		},
		Idfaqcategories: int32(cid),
	}); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
}

func (DeleteCategoryTask) Action(w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if err := queries.DeleteFAQCategory(r.Context(), int32(cid)); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
}

func (CreateCategoryTask) Action(w http.ResponseWriter, r *http.Request) {
	text := r.PostFormValue("cname")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if err := queries.CreateFAQCategory(r.Context(), sql.NullString{
		String: text,
		Valid:  true,
	}); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
}
