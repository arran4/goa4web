package linker

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

// AdminCategoryGrantsPage displays grants for a linker category.
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
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	data := Data{CategoryID: int32(cid), Actions: []string{"see", "view"}}
	cd.PageTitle = fmt.Sprintf("Category %d Grants", cid)

	if roles, err := cd.AllRoles(); err == nil {
		data.Roles = roles
	}

	rolesMap := make(map[int32]*db.Role)
	if data.Roles != nil {
		for _, r := range data.Roles {
			rolesMap[r.ID] = r
		}
	}

	grants, err := queries.ListGrants(r.Context())
	if err != nil {
		log.Printf("ListGrants: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	var userIDs []int32
	var relevantGrants []*db.Grant
	for _, g := range grants {
		if g.Section == "linker" && g.Item.Valid && g.Item.String == "category" && g.ItemID.Valid && g.ItemID.Int32 == int32(cid) {
			relevantGrants = append(relevantGrants, g)
			if g.UserID.Valid {
				userIDs = append(userIDs, g.UserID.Int32)
			}
		}
	}

	usersMap := make(map[int32]*db.SystemGetUsersByIDsRow)
	if len(userIDs) > 0 {
		if users, err := queries.SystemGetUsersByIDs(r.Context(), userIDs); err == nil {
			for _, u := range users {
				usersMap[u.Idusers] = u
			}
		}
	}

	for _, g := range relevantGrants {
		gi := GrantInfo{Grant: g}
		if g.UserID.Valid {
			if u, ok := usersMap[g.UserID.Int32]; ok {
				gi.Username = sql.NullString{String: u.Username.String, Valid: u.Username.Valid}
			}
		}
		if g.RoleID.Valid {
			if r, ok := rolesMap[g.RoleID.Int32]; ok {
				gi.RoleName = sql.NullString{String: r.Name, Valid: true}
			}
		}
		data.Grants = append(data.Grants, gi)
	}
	_ = LinkerAdminCategoryGrantsPageTmpl.Handle(w, r, data)
}

const LinkerAdminCategoryGrantsPageTmpl tasks.Template = "domains/linker/adminCategoryGrantsPage.gohtml"
