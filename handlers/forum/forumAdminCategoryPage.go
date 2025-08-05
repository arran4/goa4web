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

// AdminCategoryPage shows information about a single forum category and its topics.
func AdminCategoryPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	cid, err := strconv.Atoi(mux.Vars(r)["category"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	cat, err := queries.GetForumCategoryById(r.Context(), int32(cid))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Category not found"))
		return
	}
	topics, err := queries.GetForumTopicsByCategoryId(r.Context(), int32(cid))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Forum Category %d", cid)
	data := struct {
		*common.CoreData
		Category *db.Forumcategory
		Topics   []*db.Forumtopic
	}{
		CoreData: cd,
		Category: cat,
		Topics:   topics,
	}
	handlers.TemplateHandler(w, r, "forumAdminCategoryPage.gohtml", data)
}
