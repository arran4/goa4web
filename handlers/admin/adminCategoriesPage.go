package admin

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

func AdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Section           string
		ForumCategories   []*db.GetAllForumCategoriesWithSubcategoryCountRow
		WritingCategories []*db.WritingCategory
		LinkerCategories  []*db.GetLinkerCategoryLinkCountsRow
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
		Section:  r.URL.Query().Get("section"),
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

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
		rows, err := data.CoreData.WritingCategories()
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("adminCategories writings: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		data.WritingCategories = rows
	}
	if data.Section == "" || data.Section == "linker" {
		rows, err := data.LinkerCategoryCounts()
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("adminCategories linker: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		data.LinkerCategories = rows
	}

	handlers.TemplateHandler(w, r, "categoriesPage.gohtml", data)
}
