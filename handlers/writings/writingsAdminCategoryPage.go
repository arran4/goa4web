package writings

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

// AdminCategoryPage shows a single writing category and its writings.
func AdminCategoryPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	cid, err := strconv.Atoi(mux.Vars(r)["category"])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	cat, err := queries.GetWritingCategoryById(r.Context(), int32(cid))
	if err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}
	writings, err := queries.AdminGetWritingsByCategoryId(r.Context(), int32(cid))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	cd.PageTitle = fmt.Sprintf("Writing Category %d", cid)
	data := struct {
		*common.CoreData
		Category *db.WritingCategory
		Writings []*db.AdminGetWritingsByCategoryIdRow
	}{
		CoreData: cd,
		Category: cat,
		Writings: writings,
	}
	handlers.TemplateHandler(w, r, "writingsAdminCategoryPage.gohtml", data)
}
