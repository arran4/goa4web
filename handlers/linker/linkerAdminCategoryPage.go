package linker

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

// AdminCategoryPage shows a linker category with its links.
func AdminCategoryPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cid, err := strconv.Atoi(mux.Vars(r)["category"])
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Bad Request"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Linker Category %d", cid)
	data := struct {
		CategoryID int32
	}{CategoryID: int32(cid)}
	handlers.TemplateHandler(w, r, "linkerAdminCategoryPage.gohtml", data)
}
