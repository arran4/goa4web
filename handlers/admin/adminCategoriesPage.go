package admin

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

func AdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Section           string
		ForumCategories   []*db.GetAllForumCategoriesWithSubcategoryCountRow
		WritingCategories []*db.Writingcategory
		LinkerCategories  []*db.GetLinkerCategoryLinkCountsRow
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
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

	if err := templates.RenderTemplate(w, "categoriesPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
