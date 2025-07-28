package writings

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

func CategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Categories          []*db.WritingCategory
		CategoryBreadcrumbs []*db.WritingCategory
		Abstracts           []*db.GetPublicWritingsInCategoryForUserRow
		WritingCategoryID   int32
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	handlers.SetPageTitle(r, "Writing Categories")
	data := Data{}
	data.WritingCategoryID = 0

	categoryRows, err := cd.VisibleWritingCategories(cd.UserID)
	if err != nil {
		log.Printf("writingCategories: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writingsRows, err := cd.PublicWritings(0, r)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getPublicWritingsInCategory Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	categoryMap := map[int32]*db.WritingCategory{}
	for _, cat := range categoryRows {
		if !cd.HasGrant("writing", "category", "see", cat.Idwritingcategory) {
			continue
		}
		categoryMap[cat.Idwritingcategory] = cat
		if cat.WritingCategoryID == 0 {
			data.Categories = append(data.Categories, cat)
		}
	}

	for _, wrow := range writingsRows {
		if !cd.HasGrant("writing", "article", "see", wrow.Idwriting) {
			continue
		}
		data.Abstracts = append(data.Abstracts, wrow)
	}

	handlers.TemplateHandler(w, r, "writingsCategoriesPage.gohtml", data)
}
