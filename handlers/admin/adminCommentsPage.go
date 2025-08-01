package admin

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminCommentsPage lists recent comments.
func AdminCommentsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Comments"
	queries := cd.Queries()
	rows, err := queries.ListAllCommentsWithThreadInfoForAdmin(r.Context(), db.ListAllCommentsWithThreadInfoForAdminParams{Limit: 50, Offset: 0})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := struct {
		*common.CoreData
		Comments []*db.ListAllCommentsWithThreadInfoForAdminRow
	}{cd, rows}
	handlers.TemplateHandler(w, r, "commentsPage.gohtml", data)
}
