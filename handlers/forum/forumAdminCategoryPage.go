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
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	cat, err := queries.GetForumCategoryById(r.Context(), int32(cid))
	if err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}
	topics, err := queries.GetForumTopicsByCategoryId(r.Context(), int32(cid))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
