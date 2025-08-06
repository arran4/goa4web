package linker

import (
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

// AdminCategoryEditPage shows an edit form for a linker category.
func AdminCategoryEditPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cid, err := strconv.Atoi(mux.Vars(r)["category"])
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	cd.PageTitle = "Edit Linker Category " + strconv.Itoa(cid)
	data := struct {
		CategoryID int32
	}{CategoryID: int32(cid)}
	handlers.TemplateHandler(w, r, "linkerAdminCategoryEditPage.gohtml", data)
}

// AdminCategoryLinksPage lists links for a linker category.
func AdminCategoryLinksPage(w http.ResponseWriter, r *http.Request) {
	AdminCategoryPage(w, r)
}
