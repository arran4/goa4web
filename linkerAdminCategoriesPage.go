package goa4web

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
)

func linkerAdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Categories []*GetLinkerCategoryLinkCountsRow
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	categoryRows, err := queries.GetLinkerCategoryLinkCounts(r.Context())
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

	CustomLinkerIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "categoriesPage.gohtml", data, NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func linkerAdminCategoriesUpdatePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	cid, _ := strconv.Atoi(r.PostFormValue("cid"))
	title := r.PostFormValue("title")
	pos, _ := strconv.Atoi(r.PostFormValue("position"))
	if err := queries.RenameLinkerCategory(r.Context(), RenameLinkerCategoryParams{
		Title:            sql.NullString{Valid: true, String: title},
		Position:         int32(pos),
		Idlinkercategory: int32(cid),
	}); err != nil {
		log.Printf("renameLinkerCategory Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	order, _ := strconv.Atoi(r.PostFormValue("order"))
	if err := queries.UpdateLinkerCategorySortOrder(r.Context(), UpdateLinkerCategorySortOrderParams{
		Sortorder:        int32(order),
		Idlinkercategory: int32(cid),
	}); err != nil {
		log.Printf("updateLinkerCategorySortOrder Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	taskDoneAutoRefreshPage(w, r)
}

func linkerAdminCategoriesRenamePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	cid, _ := strconv.Atoi(r.PostFormValue("cid"))
	title := r.PostFormValue("title")
	pos, _ := strconv.Atoi(r.PostFormValue("position"))
	if err := queries.RenameLinkerCategory(r.Context(), RenameLinkerCategoryParams{
		Title:            sql.NullString{Valid: true, String: title},
		Position:         int32(pos),
		Idlinkercategory: int32(cid),
	}); err != nil {
		log.Printf("renameLinkerCategory Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	taskDoneAutoRefreshPage(w, r)
}

func linkerAdminCategoriesDeletePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	cid, _ := strconv.Atoi(r.PostFormValue("cid"))
	rows, _ := queries.GetLinkerCategoryLinkCounts(r.Context())
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
	taskDoneAutoRefreshPage(w, r)
}

func linkerAdminCategoriesCreatePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	title := r.PostFormValue("title")
	rows, _ := queries.GetLinkerCategoryLinkCounts(r.Context())
	pos := len(rows) + 1
	if err := queries.CreateLinkerCategory(r.Context(), CreateLinkerCategoryParams{
		Title:    sql.NullString{Valid: true, String: title},
		Position: int32(pos),
	}); err != nil {
		log.Printf("renameLinkerCategory Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	taskDoneAutoRefreshPage(w, r)
}
