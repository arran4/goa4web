package admin

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Section           string
		ForumCategories   []*db.GetAllForumCategoriesWithSubcategoryCountRow
		WritingCategories []*db.WritingCategory
		LinkerCategories  []*db.GetLinkerCategoryLinkCountsRow
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	handlers.SetPageTitle(r, "Admin Categories")
	data := Data{
		CoreData: cd,
		Section:  r.URL.Query().Get("section"),
	}
	queries := cd.Queries()

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

	handlers.TemplateHandler(w, r, "siteAdminCategoriesPage.gohtml", data)
}
