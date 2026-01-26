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

// adminUserGrantsPage shows direct grants for a user and allows editing.
func adminUserGrantsPage(w http.ResponseWriter, r *http.Request) {
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
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
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

	AdminUserGrantsPageTmpl.Handle(w, r, data)
}

const AdminUserGrantsPageTmpl tasks.Template = "admin/userGrantsPage.gohtml"
