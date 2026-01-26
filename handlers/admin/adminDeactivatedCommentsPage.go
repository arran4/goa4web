package admin

import (
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminDeactivatedCommentsPage lists comments that have been deactivated.
func AdminDeactivatedCommentsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Deactivated Comments"
	rows, err := cd.Queries().AdminListDeactivatedComments(r.Context(), db.AdminListDeactivatedCommentsParams{Limit: 50, Offset: 0})
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data := struct {
		*common.CoreData
		Comments []*db.AdminListDeactivatedCommentsRow
	}{cd, rows}
	AdminDeactivatedCommentsPageTmpl.Handle(w, r, data)
}

const AdminDeactivatedCommentsPageTmpl tasks.Template = "admin/deactivatedCommentsPage.gohtml"
