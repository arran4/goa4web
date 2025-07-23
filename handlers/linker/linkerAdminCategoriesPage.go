package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

func AdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Categories []*db.GetLinkerCategoryLinkCountsRow
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}

	categoryRows, err := data.LinkerCategoryCounts()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("adminCategories Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Categories = categoryRows

	handlers.TemplateHandler(w, r, "categoriesPage.gohtml", data)
}

type updateCategoryTask struct{ tasks.TaskString }

var UpdateCategoryTask = &updateCategoryTask{TaskString: TaskUpdate}
var _ tasks.Task = (*updateCategoryTask)(nil)

func (updateCategoryTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cid, _ := strconv.Atoi(r.PostFormValue("cid"))
	title := r.PostFormValue("title")
	pos, _ := strconv.Atoi(r.PostFormValue("position"))
	if err := queries.RenameLinkerCategory(r.Context(), db.RenameLinkerCategoryParams{
		Title:            sql.NullString{Valid: true, String: title},
		Position:         int32(pos),
		Idlinkercategory: int32(cid),
	}); err != nil {
		return fmt.Errorf("rename linker category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	order, _ := strconv.Atoi(r.PostFormValue("order"))
	if err := queries.UpdateLinkerCategorySortOrder(r.Context(), db.UpdateLinkerCategorySortOrderParams{
		Sortorder:        int32(order),
		Idlinkercategory: int32(cid),
	}); err != nil {
		return fmt.Errorf("update linker category sort order fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}

type renameCategoryTask struct{ tasks.TaskString }

var RenameCategoryTask = &renameCategoryTask{TaskString: TaskRenameCategory}
var _ tasks.Task = (*renameCategoryTask)(nil)

func (renameCategoryTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cid, _ := strconv.Atoi(r.PostFormValue("cid"))
	title := r.PostFormValue("title")
	pos, _ := strconv.Atoi(r.PostFormValue("position"))
	if err := queries.RenameLinkerCategory(r.Context(), db.RenameLinkerCategoryParams{
		Title:            sql.NullString{Valid: true, String: title},
		Position:         int32(pos),
		Idlinkercategory: int32(cid),
	}); err != nil {
		return fmt.Errorf("rename linker category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}

type deleteCategoryTask struct{ tasks.TaskString }

var DeleteCategoryTask = &deleteCategoryTask{TaskString: TaskDeleteCategory}
var _ tasks.Task = (*deleteCategoryTask)(nil)

func (deleteCategoryTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cid, _ := strconv.Atoi(r.PostFormValue("cid"))
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	rows, _ := cd.LinkerCategoryCounts()
	for _, c := range rows {
		if int(c.Idlinkercategory) == cid && c.Linkcount > 0 {
			http.Error(w, "Category in use", http.StatusBadRequest)
			return nil
		}
	}
	count, err := queries.CountLinksByCategory(r.Context(), int32(cid))
	if err != nil {
		return fmt.Errorf("count links fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if count > 0 {
		http.Error(w, "Category in use", http.StatusBadRequest)
		return nil
	}
	if err := queries.DeleteLinkerCategory(r.Context(), int32(cid)); err != nil {
		return fmt.Errorf("delete linker category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}

type createCategoryTask struct{ tasks.TaskString }

var CreateCategoryTask = &createCategoryTask{TaskString: TaskCreateCategory}
var _ tasks.Task = (*createCategoryTask)(nil)

func (createCategoryTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	title := r.PostFormValue("title")
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	rows, _ := cd.LinkerCategoryCounts()
	pos := len(rows) + 1
	if err := queries.CreateLinkerCategory(r.Context(), db.CreateLinkerCategoryParams{
		Title:    sql.NullString{Valid: true, String: title},
		Position: int32(pos),
	}); err != nil {
		return fmt.Errorf("create linker category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}
