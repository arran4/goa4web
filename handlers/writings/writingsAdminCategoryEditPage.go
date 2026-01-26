package writings

import (
	"database/sql"
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

// AdminCategoryEditPage shows a form to edit a single writing category.
func AdminCategoryEditPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	cid, err := strconv.Atoi(mux.Vars(r)["category"])
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	cat, err := queries.GetWritingCategoryById(r.Context(), int32(cid))
	if err != nil {
		if err == sql.ErrNoRows {
			handlers.RenderErrorPage(w, r, fmt.Errorf("Category not found"))
		} else {
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		}
		return
	}
	all, err := cd.WritingCategories()
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Edit Category %d", cid)
	data := struct {
		Category   *db.WritingCategory
		Categories []*db.WritingCategory
	}{
		Category:   cat,
		Categories: all,
	}
	WritingsAdminCategoryEditPageTmpl.Handle(w, r, data)
}

const WritingsAdminCategoryEditPageTmpl tasks.Template = "writings/writingsAdminCategoryEditPage.gohtml"
