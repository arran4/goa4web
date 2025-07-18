package linker

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

func AdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		Categories []*db.GetLinkerCategoryLinkCountsRow
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
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

	common.TemplateHandler(w, r, "categoriesPage.gohtml", data)
}

func AdminCategoriesUpdatePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	cid, _ := strconv.Atoi(r.PostFormValue("cid"))
	title := r.PostFormValue("title")
	pos, _ := strconv.Atoi(r.PostFormValue("position"))
	if err := queries.RenameLinkerCategory(r.Context(), db.RenameLinkerCategoryParams{
		Title:            sql.NullString{Valid: true, String: title},
		Position:         int32(pos),
		Idlinkercategory: int32(cid),
	}); err != nil {
		log.Printf("renameLinkerCategory Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	order, _ := strconv.Atoi(r.PostFormValue("order"))
	if err := queries.UpdateLinkerCategorySortOrder(r.Context(), db.UpdateLinkerCategorySortOrderParams{
		Sortorder:        int32(order),
		Idlinkercategory: int32(cid),
	}); err != nil {
		log.Printf("updateLinkerCategorySortOrder Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	common.TaskDoneAutoRefreshPage(w, r)
}

func AdminCategoriesRenamePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	cid, _ := strconv.Atoi(r.PostFormValue("cid"))
	title := r.PostFormValue("title")
	pos, _ := strconv.Atoi(r.PostFormValue("position"))
	if err := queries.RenameLinkerCategory(r.Context(), db.RenameLinkerCategoryParams{
		Title:            sql.NullString{Valid: true, String: title},
		Position:         int32(pos),
		Idlinkercategory: int32(cid),
	}); err != nil {
		log.Printf("renameLinkerCategory Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	common.TaskDoneAutoRefreshPage(w, r)
}

func AdminCategoriesDeletePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	cid, _ := strconv.Atoi(r.PostFormValue("cid"))
	cd := r.Context().Value(common.KeyCoreData).(*corecommon.CoreData)
	rows, _ := cd.LinkerCategoryCounts()
	for _, c := range rows {
		if int(c.Idlinkercategory) == cid && c.Linkcount > 0 {
			http.Error(w, "Category in use", http.StatusBadRequest)
			return
		}
	}
	count, err := queries.CountLinksByCategory(r.Context(), int32(cid))
	if err != nil {
		log.Printf("countLinks Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "Category in use", http.StatusBadRequest)
		return
	}
	if err := queries.DeleteLinkerCategory(r.Context(), int32(cid)); err != nil {
		log.Printf("renameLinkerCategory Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	common.TaskDoneAutoRefreshPage(w, r)
}

func AdminCategoriesCreatePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	title := r.PostFormValue("title")
	cd := r.Context().Value(common.KeyCoreData).(*corecommon.CoreData)
	rows, _ := cd.LinkerCategoryCounts()
	pos := len(rows) + 1
	if err := queries.CreateLinkerCategory(r.Context(), db.CreateLinkerCategoryParams{
		Title:    sql.NullString{Valid: true, String: title},
		Position: int32(pos),
	}); err != nil {
		log.Printf("renameLinkerCategory Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	common.TaskDoneAutoRefreshPage(w, r)
}
