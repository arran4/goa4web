package admin

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminRolePage struct {
	RoleName string
	RoleID   int32
	Data     any
}

func (p *AdminRolePage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return fmt.Sprintf("Role %s", p.RoleName), "", &AdminRolesPageTask{}
}

func (p *AdminRolePage) PageTitle() string {
	return fmt.Sprintf("Role: %s", p.RoleName)
}

func (p *AdminRolePage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	AdminRolePageTmpl.Handler(p.Data).ServeHTTP(w, r)
}

type AdminRoleTask struct{}

func (t *AdminRoleTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	role, err := cd.SelectedRole()
	if err != nil || role == nil {
		return fmt.Errorf("role not found")
	}

	id := cd.SelectedRoleID()
	emailRows, err := queries.GetVerifiedUserEmails(r.Context())
	if err != nil && !isNoRows(err) {
		return err
	}
	emailsByUser := make(map[int32][]string)
	for _, row := range emailRows {
		emailsByUser[row.UserID] = append(emailsByUser[row.UserID], row.Email)
	}

	users, err := queries.AdminListUsersByRoleID(r.Context(), id)
	if err != nil && !isNoRows(err) {
		return err
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
		return err
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

	return &AdminRolePage{
		RoleName: role.Name,
		RoleID:   role.ID,
		Data:     data,
	}
}

func isNoRows(err error) bool {
	return err == sql.ErrNoRows
}

type roleUser struct {
	ID     int32
	Email  []string
	User   sql.NullString
	UserID int32
}

func (ru *roleUser) EmailList() string {
	return strings.Join(ru.Email, ", ")
}

const AdminRolePageTmpl tasks.Template = "admin/adminRolePage.gohtml"
