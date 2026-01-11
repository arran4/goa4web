package linker

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

// AdminLinkGrantsPage displays grants for a linker item.
func AdminLinkGrantsPage(w http.ResponseWriter, r *http.Request) {
	type GrantInfo struct {
		*db.Grant
		Username sql.NullString
		RoleName sql.NullString
	}
	type Data struct {
		LinkID  int32
		Grants  []GrantInfo
		Roles   []*db.Role
		Actions []string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	lid, err := strconv.Atoi(mux.Vars(r)["link"])
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	data := Data{LinkID: int32(lid), Actions: []string{"see", "view", "reply"}}
	cd.PageTitle = fmt.Sprintf("Link %d Grants", lid)
	if roles, err := cd.AllRoles(); err == nil {
		data.Roles = roles
	}
	grants, err := queries.ListGrants(r.Context())
	if err != nil {
		log.Printf("ListGrants: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	for _, g := range grants {
		if g.Section == "linker" && g.Item.Valid && g.Item.String == "link" && g.ItemID.Valid && g.ItemID.Int32 == int32(lid) {
			gi := GrantInfo{Grant: g}
			if g.UserID.Valid {
				if u, err := queries.SystemGetUserByID(r.Context(), g.UserID.Int32); err == nil {
					gi.Username = sql.NullString{String: u.Username.String, Valid: true}
				}
			}
			if g.RoleID.Valid && data.Roles != nil {
				for _, r := range data.Roles {
					if r.ID == g.RoleID.Int32 {
						gi.RoleName = sql.NullString{String: r.Name, Valid: true}
						break
					}
				}
			}
			data.Grants = append(data.Grants, gi)
		}
	}
	LinkerAdminLinkGrantsPageTmpl.Handle(w, r, data)
}

const LinkerAdminLinkGrantsPageTmpl handlers.Page = "linker/adminLinkGrantsPage.gohtml"
