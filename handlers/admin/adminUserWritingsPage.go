package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminUserWritingsPage struct{}

func (p *AdminUserWritingsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	cpu := cd.CurrentProfileUser()
	if cpu.Idusers == 0 {
		handlers.RenderErrorPage(w, r, fmt.Errorf("user not found"))
		return
	}
	user := cd.CurrentProfileUser()
	if user == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("user not found"))
		return
	}
	queries := cd.Queries()
	rows, err := queries.AdminGetAllWritingsByAuthor(r.Context(), cpu.Idusers)
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	cd.PageTitle = fmt.Sprintf("Writings by %s", user.Username.String)
	data := struct {
		User     *db.User
		Writings []*db.AdminGetAllWritingsByAuthorRow
	}{
		User:     &db.User{Idusers: cpu.Idusers, Username: user.Username},
		Writings: rows,
	}
	AdminUserWritingsPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminUserWritingsPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Writings", "", &AdminUserProfilePage{}
}

func (p *AdminUserWritingsPage) PageTitle() string {
	return "User Writings"
}

var _ common.Page = (*AdminUserWritingsPage)(nil)
var _ http.Handler = (*AdminUserWritingsPage)(nil)

const AdminUserWritingsPageTmpl tasks.Template = "admin/userWritingsPage.gohtml"
