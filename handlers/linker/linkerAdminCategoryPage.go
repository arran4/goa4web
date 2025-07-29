package linker

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

// AdminCategoryPage shows a linker category with its links.
func AdminCategoryPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	cid, err := strconv.Atoi(mux.Vars(r)["category"])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	cat, err := queries.GetLinkerCategoryById(r.Context(), int32(cid))
	if err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}
	links, err := queries.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending(r.Context(), db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingParams{Idlinkercategory: int32(cid)})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	cd.PageTitle = fmt.Sprintf("Linker Category %d", cid)
	data := struct {
		*common.CoreData
		Category *db.LinkerCategory
		Links    []*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow
	}{
		CoreData: cd,
		Category: cat,
		Links:    links,
	}
	handlers.TemplateHandler(w, r, "linkerAdminCategoryPage.gohtml", data)
}
