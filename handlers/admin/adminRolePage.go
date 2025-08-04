package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// adminRolePage shows details for a role including grants and users.
func adminRolePage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()

	role, err := queries.AdminGetRoleByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "role not found", http.StatusNotFound)
		return
	}
	cd.PageTitle = fmt.Sprintf("Role %s", role.Name)

	users, err := queries.AdminListUsersByRoleID(r.Context(), int32(id))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	groups, err := buildGrantGroups(r.Context(), cd, int32(id))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
