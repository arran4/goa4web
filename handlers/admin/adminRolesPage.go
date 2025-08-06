package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminRolesPage lists all roles with their public profile access flag.
func AdminRolesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Roles []*db.Role
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin Roles"
	roles, err := cd.AllRoles()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list roles: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data := Data{Roles: roles}
	handlers.TemplateHandler(w, r, "rolesPage.gohtml", data)
}
