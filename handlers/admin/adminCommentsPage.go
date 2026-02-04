package admin

import (
	"github.com/arran4/goa4web/internal/tasks"
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
	rows, err := queries.AdminListAllCommentsWithThreadInfo(r.Context(), db.AdminListAllCommentsWithThreadInfoParams{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	data := struct {
		*common.CoreData
		Comments []*db.AdminListAllCommentsWithThreadInfoRow
	}{cd, rows}
	AdminCommentsPageTmpl.Handle(w, r, data)
}

const AdminCommentsPageTmpl tasks.Template = "admin/commentsPage.gohtml"
