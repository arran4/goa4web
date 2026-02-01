package blogs

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// AdminPage shows the blog administration index with a list of blogs.
func AdminPage(w http.ResponseWriter, r *http.Request) {
	type BlogRoleRow struct {
		Idusers  int32
		Username sql.NullString
		Email    string
		Roles    sql.NullString
	}
	type Data struct {
		Rows []BlogRoleRow
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	offset := cd.Offset()
	ps := cd.PageSize()
	cd.NextLink = fmt.Sprintf("/admin/blogs?offset=%d", offset+ps)
	if offset > 0 {
		cd.PrevLink = fmt.Sprintf("/admin/blogs?offset=%d", offset-ps)
		cd.StartLink = "/admin/blogs?offset=0"
	}
	cd.PageTitle = "Blog Admin"

	roles, err := cd.AllRoles()
	if err != nil {
		log.Printf("AllRoles: %v", err)
	}
	grants, err := cd.Queries().ListGrants(r.Context())
	if err != nil {
		log.Printf("ListGrants: %v", err)
	}
	userRolesMap := make(map[int32][]string)
	roleUsersMap := make(map[int32][]int32)
	for _, role := range roles {
		uids, err := cd.Queries().AdminListUserIDsByRole(r.Context(), role.Name)
		if err != nil {
			log.Printf("AdminListUserIDsByRole %s: %v", role.Name, err)
			continue
		}
		roleUsersMap[role.ID] = uids
		for _, uid := range uids {
			userRolesMap[uid] = append(userRolesMap[uid], role.Name)
		}
	}

	relevantUserIDs := make(map[int32]bool)
	for _, g := range grants {
		if g.Section == "blogs" {
			if g.UserID.Valid {
				relevantUserIDs[g.UserID.Int32] = true
			}
			if g.RoleID.Valid {
				if uids, ok := roleUsersMap[g.RoleID.Int32]; ok {
					for _, uid := range uids {
						relevantUserIDs[uid] = true
					}
				}
			}
		}
	}

	var data Data
	for uid := range relevantUserIDs {
		u, err := cd.Queries().SystemGetUserByID(r.Context(), uid)
		if err != nil {
			log.Printf("SystemGetUserByID %d: %v", uid, err)
			continue
		}
		rolesList := userRolesMap[uid]
		sort.Strings(rolesList)
		data.Rows = append(data.Rows, BlogRoleRow{
			Idusers:  u.Idusers,
			Username: u.Username,
			Email:    u.Email.String,
			Roles: sql.NullString{
				String: strings.Join(rolesList, ", "),
				Valid:  len(rolesList) > 0,
			},
		})
	}
	sort.Slice(data.Rows, func(i, j int) bool {
		return data.Rows[i].Idusers < data.Rows[j].Idusers
	})

	BlogsAdminPageTmpl.Handle(w, r, data)
}

const BlogsAdminPageTmpl tasks.Template = "blogs/adminPage.gohtml"
