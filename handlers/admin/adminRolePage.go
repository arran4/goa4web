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

	role, err := queries.GetRoleByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "role not found", http.StatusNotFound)
		return
	}
	cd.PageTitle = fmt.Sprintf("Role %s", role.Name)

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

	type GrantInfo struct {
		*db.Grant
		Link string
	}
	var ginfos []GrantInfo
	for _, g := range grants {
		gi := GrantInfo{Grant: g}
		if g.Item.Valid && g.ItemID.Valid {
			switch g.Section {
			case "forum":
				switch g.Item.String {
				case "topic":
					gi.Link = fmt.Sprintf("/admin/forum/topic/%d/grants#g%d", g.ItemID.Int32, g.ID)
				case "category":
					gi.Link = fmt.Sprintf("/admin/forum/category/%d/grants#g%d", g.ItemID.Int32, g.ID)
				}
			case "linker":
				if g.Item.String == "category" {
					gi.Link = fmt.Sprintf("/admin/linker/category/%d/grants#g%d", g.ItemID.Int32, g.ID)
				}
			case "writings":
				if g.Item.String == "category" {
					gi.Link = fmt.Sprintf("/admin/writings/category/%d/permissions#g%d", g.ItemID.Int32, g.ID)
				}
			}
		} else if g.Section == "role" && g.Action != "" {
			if roles, err := cd.AllRoles(); err == nil {
				for _, ro := range roles {
					if ro.Name == g.Action {
						gi.Link = fmt.Sprintf("/admin/role/%d#g%d", ro.ID, g.ID)
						break
					}
				}
			}
		}
		ginfos = append(ginfos, gi)
	}

	data := struct {
		*common.CoreData
		Role   *db.Role
		Users  []*db.ListUsersByRoleIDRow
		Grants []GrantInfo
	}{
		CoreData: cd,
		Role:     role,
		Users:    users,
		Grants:   ginfos,
	}

	handlers.TemplateHandler(w, r, "adminRolePage.gohtml", data)
}
