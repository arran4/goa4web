package admin

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminDeactivatedCommentsPage struct{}

func (p *AdminDeactivatedCommentsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Deactivated Comments"
	rows, err := cd.Queries().AdminListDeactivatedComments(r.Context(), db.AdminListDeactivatedCommentsParams{Limit: 50, Offset: 0})
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	data := struct {
		*common.CoreData
		Comments []*db.AdminListDeactivatedCommentsRow
	}{cd, rows}
	AdminDeactivatedCommentsPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminDeactivatedCommentsPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Deactivated Comments", "/admin/comments/deactivated", &AdminCommentsPage{}
}

func (p *AdminDeactivatedCommentsPage) PageTitle() string {
	return "Deactivated Comments"
}

var _ common.Page = (*AdminDeactivatedCommentsPage)(nil)
var _ http.Handler = (*AdminDeactivatedCommentsPage)(nil)

const AdminDeactivatedCommentsPageTmpl tasks.Template = "admin/deactivatedCommentsPage.gohtml"
