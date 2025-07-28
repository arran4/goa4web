package admin

import (
	"database/sql"
	"errors"
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

	role, err := queries.GetRoleByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "role not found", http.StatusNotFound)
		return
	}

	users, err := queries.ListUsersByRoleID(r.Context(), int32(id))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	grants, err := queries.ListGrantsByRoleID(r.Context(), sql.NullInt32{Int32: int32(id), Valid: true})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		*common.CoreData
		Role   *db.Role
		Users  []*db.ListUsersByRoleIDRow
		Grants []*db.Grant
	}{
		CoreData: cd,
		Role:     role,
		Users:    users,
		Grants:   grants,
	}

	handlers.TemplateHandler(w, r, "adminRolePage.gohtml", data)
}
