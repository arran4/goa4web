package writings

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

// AdminCategoryPage shows a single writing category and its writings.
func AdminCategoryPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	cid, err := strconv.Atoi(mux.Vars(r)["category"])
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	cat, err := queries.GetWritingCategoryById(r.Context(), int32(cid))
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Category not found"))
		return
	}
	writings, err := queries.AdminGetWritingsByCategoryId(r.Context(), int32(cid))
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	cd.PageTitle = fmt.Sprintf("Writing Category %d", cid)
	data := struct {
		Category *db.WritingCategory
		Writings []*db.AdminGetWritingsByCategoryIdRow
	}{
		Category: cat,
		Writings: writings,
	}
	WritingsAdminCategoryPageTmpl.Handle(w, r, data)
}

const WritingsAdminCategoryPageTmpl tasks.Template = "writings/writingsAdminCategoryPage.gohtml"
