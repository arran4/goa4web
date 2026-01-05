package forum

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

// AdminCategoryGrantsPage displays grants for a forum category.
func AdminCategoryGrantsPage(w http.ResponseWriter, r *http.Request) {
	type GrantInfo struct {
		*db.Grant
		Username sql.NullString
		RoleName sql.NullString
	}
	type Data struct {
		CategoryID int32
		Grants     []GrantInfo
		Roles      []*db.Role
		Actions    []string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	cid, err := strconv.Atoi(mux.Vars(r)["category"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	cd.PageTitle = fmt.Sprintf("Forum - Category %d Grants", cid)
	data := Data{CategoryID: int32(cid), Actions: []string{"see", "view"}}
	if roles, err := cd.AllRoles(); err == nil {
		data.Roles = roles
	}
	grants, err := queries.ListGrants(r.Context())
	if err != nil {
		log.Printf("ListGrants: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	for _, g := range grants {
		if g.Section == "forum" && g.Item.Valid && g.Item.String == "category" && g.ItemID.Valid && g.ItemID.Int32 == int32(cid) {
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
	handlers.TemplateHandler(w, r, ForumAdminCategoryGrantsPageTmpl, data)
}
