package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

type roleUser struct {
	ID     int32
	Email  []string
	User   sql.NullString
	UserID int32
}

func (ru *roleUser) EmailList() string {
	return strings.Join(ru.Email, ", ")
}

// adminRolePage shows details for a role including grants and users.
func adminRolePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	role, err := cd.SelectedRole()
	if err != nil || role == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("role not found"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Role: %s", role.Name)
	cd.SetCurrentPage(&AdminRolePageBreadcrumb{RoleName: role.Name, RoleID: role.ID})

	id := cd.SelectedRoleID()
	emailRows, err := queries.GetVerifiedUserEmails(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	emailsByUser := make(map[int32][]string)
	for _, row := range emailRows {
		emailsByUser[row.UserID] = append(emailsByUser[row.UserID], row.Email)
	}

	users, err := queries.AdminListUsersByRoleID(r.Context(), id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	roleUsers := make([]*roleUser, 0, len(users))
	for _, u := range users {
		ru := &roleUser{ID: u.Idusers, User: u.Username, UserID: u.Idusers}
		if emails, ok := emailsByUser[u.Idusers]; ok {
			ru.Email = emails
		}
		roleUsers = append(roleUsers, ru)
	}

	groups, err := buildGrantGroups(r.Context(), cd, id)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	data := struct {
		Role        *db.Role
		Users       []*roleUser
		GrantGroups []GrantGroup
	}{
		Role:        role,
		Users:       roleUsers,
		GrantGroups: groups,
	}

	AdminRolePageTmpl.Handle(w, r, data)
}

const AdminRolePageTmpl tasks.Template = "admin/adminRolePage.gohtml"
