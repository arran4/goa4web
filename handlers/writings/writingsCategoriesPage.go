package writings

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
)

func CategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		Categories          []*db.WritingCategory
		CategoryBreadcrumbs []*db.WritingCategory
		EditingCategoryId   int32
		IsAdmin             bool
		IsWriter            bool
		Abstracts           []*db.GetPublicWritingsInCategoryForUserRow
		WritingCategoryID   int32
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
	}

	data.IsAdmin = data.CoreData.HasRole("administrator") && data.CoreData.AdminMode
	data.IsWriter = data.CoreData.HasRole("content writer") || data.IsAdmin
	editID, _ := strconv.Atoi(r.URL.Query().Get("edit"))
	data.EditingCategoryId = int32(editID)
	data.WritingCategoryID = 0

	categoryRows, err := data.CoreData.WritingCategories()
	if err != nil {
		log.Printf("writingCategories: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writingsRows, err := data.CoreData.PublicWritings(0, r)
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
		if !data.CoreData.HasGrant("writing", "category", "see", cat.Idwritingcategory) {
			continue
		}
		categoryMap[cat.Idwritingcategory] = cat
		if cat.WritingCategoryID == 0 {
			data.Categories = append(data.Categories, cat)
		}
	}

	for _, wrow := range writingsRows {
		if !data.CoreData.HasGrant("writing", "article", "see", wrow.Idwriting) {
			continue
		}
		data.Abstracts = append(data.Abstracts, wrow)
	}

	if err := templates.RenderTemplate(w, "categoriesPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
