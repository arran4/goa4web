package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminUserGrantsPage struct{}

func (p *AdminUserGrantsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	user := cd.CurrentProfileUser()
	if user == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("user not found"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Grants: %s", user.Username.String)

	roles := cd.CurrentProfileRoles()
	groups, err := buildGrantGroupsForUser(r.Context(), cd, user.Idusers)
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	data := struct {
		User        *db.SystemGetUserByIDRow
		Roles       []*db.GetPermissionsByUserIDRow
		GrantGroups []GrantGroup
	}{
		User:        user,
		Roles:       roles,
		GrantGroups: groups,
	}

	AdminUserGrantsPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminUserGrantsPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Grants", "", &AdminUserProfilePage{}
}

func (p *AdminUserGrantsPage) PageTitle() string {
	return "User Grants"
}

var _ common.Page = (*AdminUserGrantsPage)(nil)
var _ http.Handler = (*AdminUserGrantsPage)(nil)

const AdminUserGrantsPageTmpl tasks.Template = "admin/userGrantsPage.gohtml"
