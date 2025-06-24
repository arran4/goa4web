package goa4web

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
)

func adminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Section           string
		ForumCategories   []*GetAllForumCategoriesWithSubcategoryCountRow
		WritingCategories []*Writingcategory
		LinkerCategories  []*GetLinkerCategoryLinkCountsRow
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Section:  r.URL.Query().Get("section"),
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	if data.Section == "" || data.Section == "forum" {
		rows, err := queries.GetAllForumCategoriesWithSubcategoryCount(r.Context())
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("adminCategories forum: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		data.ForumCategories = rows
	}
	if data.Section == "" || data.Section == "writings" {
		rows, err := queries.FetchAllCategories(r.Context())
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("adminCategories writings: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		data.WritingCategories = rows
	}
	if data.Section == "" || data.Section == "linker" {
		rows, err := queries.GetLinkerCategoryLinkCounts(r.Context())
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("adminCategories linker: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		data.LinkerCategories = rows
	}

	if err := renderTemplate(w, r, "adminCategoriesPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
