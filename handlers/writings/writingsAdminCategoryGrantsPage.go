package writings

import (
	"database/sql"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

// AdminCategoryGrantsPage shows grants for a writing category.
func AdminCategoryGrantsPage(w http.ResponseWriter, r *http.Request) {
	type GrantInfo struct {
		*db.Grant
		Username sql.NullString
		RoleName sql.NullString
	}
	type Data struct {
		*common.CoreData
		CategoryID int32
		Grants     []GrantInfo
		Roles      []*db.Role
		Actions    []string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	handlers.SetPageTitle(r, "Category Grants")
	queries := cd.Queries()
	cid, err := strconv.Atoi(mux.Vars(r)["category"])
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	data := Data{CoreData: cd, CategoryID: int32(cid), Actions: []string{"see", "view", "post", "edit"}}
	if roles, err := cd.AllRoles(); err == nil {
		data.Roles = roles
	}

	grants, err := queries.ListGrants(r.Context())
	if err != nil {
		log.Printf("ListGrants: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	for _, g := range grants {
		if g.Section == "writing" && g.Item.Valid && g.Item.String == "category" && g.ItemID.Valid && g.ItemID.Int32 == int32(cid) {
			gi := GrantInfo{Grant: g}
			if g.UserID.Valid {
				if u, err := queries.GetUserById(r.Context(), g.UserID.Int32); err == nil {
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
	handlers.TemplateHandler(w, r, "adminCategoryGrantsPage.gohtml", data)

	// TODO ??? handlers.TemplateHandler(w, r, "writingsAdminCategoryGrantsPage.gohtml", data)
}
