package forum

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

// AdminCategoryEditPage displays a form to edit a forum category.
func AdminCategoryEditPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	cid, err := strconv.Atoi(mux.Vars(r)["category"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	cat, err := queries.GetForumCategoryById(r.Context(), db.GetForumCategoryByIdParams{Idforumcategory: int32(cid), ViewerID: cd.UserID})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Category not found"))
		return
	}
	cats, err := queries.GetAllForumCategories(r.Context(), db.GetAllForumCategoriesParams{ViewerID: cd.UserID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Edit Forum Category %d", cid)
	data := struct {
		Category   *db.Forumcategory
		Categories []*db.Forumcategory
	}{
		Category:   cat,
		Categories: cats,
	}
	handlers.TemplateHandler(w, r, "forumAdminCategoryEditPage.gohtml", data)
}
