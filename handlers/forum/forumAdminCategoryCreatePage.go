package forum

import (
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminCategoryCreatePage displays a form to create a new forum category.
func AdminCategoryCreatePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	cats, err := queries.GetAllForumCategories(r.Context(), db.GetAllForumCategoriesParams{ViewerID: cd.UserID})
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	parentID, _ := strconv.Atoi(r.URL.Query().Get("category"))
	cd.PageTitle = "Create Forum Category"
	data := struct {
		Categories []*db.Forumcategory
		ParentID   int
	}{
		Categories: cats,
		ParentID:   parentID,
	}
	handlers.TemplateHandler(w, r, "forumAdminCategoryCreatePage.gohtml", data)
}
