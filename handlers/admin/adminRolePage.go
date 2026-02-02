package admin

import (
	"database/sql"
	"strings"

	"github.com/arran4/goa4web/internal/tasks"
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

const AdminRolePageTmpl tasks.Template = "admin/adminRolePage.gohtml"
