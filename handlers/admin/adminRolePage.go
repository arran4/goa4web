package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// adminRolePage shows details for a role including grants and users.
func adminRolePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	role, err := cd.SelectedRole()
	if err != nil || role == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("role not found"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Role: %s", role.Name)

	id := cd.SelectedRoleID()
	users, err := queries.AdminListUsersByRoleID(r.Context(), id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	groups, err := buildGrantGroups(r.Context(), cd, id)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	data := struct {
		*common.CoreData
		Role        *db.Role
		Users       []*db.AdminListUsersByRoleIDRow
		GrantGroups []GrantGroup
	}{
		CoreData:    cd,
		Role:        role,
		Users:       users,
		GrantGroups: groups,
	}

	handlers.TemplateHandler(w, r, "adminRolePage.gohtml", data)
}
