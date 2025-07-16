package writings

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"

	"github.com/gorilla/mux"
	"golang.org/x/exp/slices"
)

func CategoryPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		Categories          []*db.WritingCategory
		CategoryBreadcrumbs []*db.WritingCategory
		EditingCategoryId   int32
		CategoryId          int32
		WritingCategoryID   int32
		IsAdmin             bool
		IsWriter            bool
		Abstracts           []*db.GetPublicWritingsInCategoryForUserRow
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
	}

	data.IsAdmin = data.CoreData.HasRole("administrator") && data.CoreData.AdminMode
	data.IsWriter = data.CoreData.HasRole("content writer") || data.IsAdmin
	editID, _ := strconv.Atoi(r.URL.Query().Get("edit"))
	data.EditingCategoryId = int32(editID)

	vars := mux.Vars(r)
	categoryId, _ := strconv.Atoi(vars["category"])
	data.CategoryId = int32(categoryId)
	data.WritingCategoryID = data.CategoryId

	categoryRows, err := data.CoreData.VisibleWritingCategories(data.CoreData.UserID)
	if err != nil {
		log.Printf("writingCategories: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	writingsRows, err := data.CoreData.PublicWritings(data.CategoryId, r)
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
		if !data.CoreData.HasGrant("writing", "article", "see", wrow.Idwriting) {
			continue
		}
		data.Abstracts = append(data.Abstracts, wrow)
	}

	common.TemplateHandler(w, r, "writingsCategoryPage", data)
}
