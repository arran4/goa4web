package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminUserListPageTask struct{}

func (t *AdminUserListPageTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Users"
	if _, err := cd.AdminListUsers(); err != nil {
		return err
	}
	return AdminUserListPageTmpl.Handler(struct{}{})
}

func (t *AdminUserListPageTask) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Users", "/admin/user", &AdminPageTask{}
}

var _ tasks.Task = (*AdminUserListPageTask)(nil)
var _ tasks.HasBreadcrumb = (*AdminUserListPageTask)(nil)

const AdminUserListPageTmpl tasks.Template = "admin/userList.gohtml"

type AdminUserProfilePageBreadcrumb struct {
	UserID   int32
	UserName string
}

func (p *AdminUserProfilePageBreadcrumb) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	title := fmt.Sprintf("User %d", p.UserID)
	if p.UserName != "" {
		title = fmt.Sprintf("User %s", p.UserName)
	}
	return title, "", &AdminUserListPageTask{}
}

type AdminUserProfilePage struct {
	UserID   int32
	UserName string
	Data     any
}

func (p *AdminUserProfilePage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	title := fmt.Sprintf("User %d", p.UserID)
	if p.UserName != "" {
		title = fmt.Sprintf("User %s", p.UserName)
	}
	return title, "", &AdminUserListPageTask{}
}

func (p *AdminUserProfilePage) PageTitle() string {
	if p.UserName != "" {
		return fmt.Sprintf("User %s", p.UserName)
	}
	return fmt.Sprintf("User %d", p.UserID)
}

func (p *AdminUserProfilePage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	AdminUserProfilePageTmpl.Handler(p.Data).ServeHTTP(w, r)
}

type AdminUserProfileTask struct{}

func (t *AdminUserProfileTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	user := cd.CurrentProfileUser()
	if user == nil {
		return handlers.ErrNotFound
	}
	return &AdminUserProfilePage{
		UserID: user.Idusers,
		UserName: user.Username.String,
		Data: struct{}{},
	}
}
