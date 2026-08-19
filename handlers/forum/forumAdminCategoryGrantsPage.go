package forum

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
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	var userIDs []int32
	var filteredGrants []GrantInfo
	for _, g := range grants {
		if g.Section == "forum" && g.Item.Valid && g.Item.String == "category" && g.ItemID.Valid && g.ItemID.Int32 == int32(cid) {
			filteredGrants = append(filteredGrants, GrantInfo{Grant: g})
			if g.UserID.Valid {
				userIDs = append(userIDs, g.UserID.Int32)
			}
		}
	}

	usersMap := make(map[int32]sql.NullString)
	if len(userIDs) > 0 {
		if users, err := queries.SystemGetUsersByIDs(r.Context(), userIDs); err == nil {
			for _, u := range users {
				usersMap[u.Idusers] = u.Username
			}
		}
	}

	rolesMap := make(map[int32]sql.NullString)
	if data.Roles != nil {
		for _, role := range data.Roles {
			rolesMap[role.ID] = sql.NullString{String: role.Name, Valid: true}
		}
	}

	for _, gi := range filteredGrants {
		if gi.UserID.Valid {
			if username, ok := usersMap[gi.UserID.Int32]; ok {
				gi.Username = username
			}
		}
		if gi.RoleID.Valid && data.Roles != nil {
			if roleName, ok := rolesMap[gi.RoleID.Int32]; ok {
				gi.RoleName = roleName
			}
		}
		data.Grants = append(data.Grants, gi)
	}

	_ = ForumAdminCategoryGrantsPageTmpl.Handle(w, r, data)
}

const ForumAdminCategoryGrantsPageTmpl tasks.Template = "domains/forum/adminCategoryGrantsPage.gohtml"
