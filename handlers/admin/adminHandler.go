package admin

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminPageTask struct{}

func (t *AdminPageTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin"
	if _, err := cd.AdminDashboardStats(); err != nil {
		return err
	}
	return AdminPageTmpl.Handler(struct{}{})
}

func (t *AdminPageTask) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Admin", "/admin", nil
}

// Ensure interface implementation
var _ tasks.Task = (*AdminPageTask)(nil)
var _ tasks.HasBreadcrumb = (*AdminPageTask)(nil)

const AdminPageTmpl tasks.Template = "admin/page.gohtml"
