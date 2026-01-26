package linker

import (
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

// AdminCategoryEditPage renders the edit form for a linker category.
func AdminCategoryEditPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cid, err := strconv.Atoi(mux.Vars(r)["category"])
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	cat, err := cd.SelectedLinkerCategory(int32(cid))
	if err != nil || cat == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("category not found"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Edit Category %d", cid)
	data := struct {
		Category *db.LinkerCategory
	}{Category: cat}
	LinkerAdminCategoryEditPageTmpl.Handle(w, r, data)
}

const LinkerAdminCategoryEditPageTmpl tasks.Template = "linker/linkerAdminCategoryEditPage.gohtml"
