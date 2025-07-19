package faq

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

type RenameCategoryTask struct{ tasks.TaskString }

var _ tasks.Task = (*RenameCategoryTask)(nil)

type DeleteCategoryTask struct{ tasks.TaskString }

var _ tasks.Task = (*DeleteCategoryTask)(nil)

type CreateCategoryTask struct{ tasks.TaskString }

var _ tasks.Task = (*CreateCategoryTask)(nil)

var renameCategoryTask = &RenameCategoryTask{TaskString: TaskRenameCategory}
var deleteCategoryTask = &DeleteCategoryTask{TaskString: TaskDeleteCategory}
var createCategoryTask = &CreateCategoryTask{TaskString: TaskCreateCategory}

func (RenameCategoryTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskRenameCategory)(r, m)
}

func (DeleteCategoryTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskDeleteCategory)(r, m)
}

func (CreateCategoryTask) Match(r *http.Request, m *mux.RouteMatch) bool {
	return tasks.HasTask(TaskCreateCategory)(r, m)
}

func AdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Rows []*db.GetFAQCategoriesWithQuestionCountRow
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	rows, err := queries.GetFAQCategoriesWithQuestionCount(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Rows = rows

	handlers.TemplateHandler(w, r, "adminCategoriesPage.gohtml", data)
}

func (RenameCategoryTask) Action(w http.ResponseWriter, r *http.Request) {
	text := r.PostFormValue("cname")
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

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
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	if err := queries.DeleteFAQCategory(r.Context(), int32(cid)); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
}

func (CreateCategoryTask) Action(w http.ResponseWriter, r *http.Request) {
	text := r.PostFormValue("cname")
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

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
