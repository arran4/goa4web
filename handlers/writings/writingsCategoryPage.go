package writings

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"

	"github.com/gorilla/mux"
)

func CategoryPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Request           *http.Request
		CategoryId        int32
		WritingCategoryID int32
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)

	vars := mux.Vars(r)
	categoryID, _ := strconv.Atoi(vars["category"])
	data := Data{
		Request:           r,
		CategoryId:        int32(categoryID),
		WritingCategoryID: int32(categoryID),
	}

	cats, err := cd.VisibleWritingCategories()
	if err != nil {
		log.Printf("writingCategories: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Category %d", data.CategoryId)
	for _, cat := range cats {
		if cat.Idwritingcategory == data.CategoryId && cat.Title.Valid {
			cd.PageTitle = fmt.Sprintf("Category: %s", cat.Title.String)
			break
		}
	}

	handlers.TemplateHandler(w, r, WritingsCategoryPageTmpl, data)
}
