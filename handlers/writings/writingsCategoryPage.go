package writings

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	"github.com/gorilla/mux"
	"golang.org/x/exp/slices"
)

func CategoryPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Categories          []*db.WritingCategory
		CategoryBreadcrumbs []*db.WritingCategory
		CategoryId          int32
		WritingCategoryID   int32
		Abstracts           []*db.GetPublicWritingsInCategoryForUserRow
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := Data{}

	vars := mux.Vars(r)
	categoryId, _ := strconv.Atoi(vars["category"])
	data.CategoryId = int32(categoryId)
	data.WritingCategoryID = data.CategoryId

	categoryRows, err := cd.VisibleWritingCategories(cd.UserID)
	if err != nil {
		log.Printf("writingCategories: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	writingsRows, err := cd.PublicWritings(data.CategoryId, r)
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
		if cat.WritingCategoryID == data.CategoryId {
			data.Categories = append(data.Categories, cat)
		}
	}
	for cid := data.CategoryId; len(data.CategoryBreadcrumbs) < len(categoryRows); {
		cat, ok := categoryMap[cid]
		if ok {
			data.CategoryBreadcrumbs = append(data.CategoryBreadcrumbs, cat)
			cid = cat.WritingCategoryID
		} else {
			break
		}
	}
	slices.Reverse(data.CategoryBreadcrumbs)
	for _, wrow := range writingsRows {
		if !cd.HasGrant("writing", "article", "see", wrow.Idwriting) {
			continue
		}
		data.Abstracts = append(data.Abstracts, wrow)
	}

	handlers.TemplateHandler(w, r, "writingsCategoryPage", data)
}
