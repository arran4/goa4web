package goa4web

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
)

func writingsCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Categories                       []*Writingcategory
		CategoryBreadcrumbs              []*Writingcategory
		EditingCategoryId                int32
		IsAdmin                          bool
		IsWriter                         bool
		Abstracts                        []*GetPublicWritingsInCategoryRow
		WritingcategoryIdwritingcategory int32
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
	}

	data.IsAdmin = data.CoreData.HasRole("administrator")
	data.IsWriter = data.CoreData.HasRole("writer") || data.IsAdmin
	editID, _ := strconv.Atoi(r.URL.Query().Get("edit"))
	data.EditingCategoryId = int32(editID)
	data.WritingcategoryIdwritingcategory = 0

	queries := r.Context().Value(common.KeyQueries).(*Queries)

	categoryRows, err := queries.FetchAllCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllWritingCategories Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	writingsRows, err := queries.GetPublicWritingsInCategory(r.Context(), GetPublicWritingsInCategoryParams{
		WritingcategoryIdwritingcategory: 0,
		Limit:                            15,
		Offset:                           0,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getPublicWritingsInCategory Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	categoryMap := map[int32]*Writingcategory{}
	for _, cat := range categoryRows {
		categoryMap[cat.Idwritingcategory] = cat
		if cat.WritingcategoryIdwritingcategory == 0 {
			data.Categories = append(data.Categories, cat)
		}
	}

	data.Abstracts = writingsRows

	CustomWritingsIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "categoriesPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
