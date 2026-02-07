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

type AdminUserImagebbsPage struct{}

func (p *AdminUserImagebbsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	rows, err := queries.GetImagePostsByUserDescendingAll(r.Context(), db.GetImagePostsByUserDescendingAllParams{
		UsersIdusers: cpu.Idusers,
		Limit:        100,
		Offset:       0,
	})
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	cd.PageTitle = fmt.Sprintf("Images by %s", user.Username.String)
	data := struct {
		User  *db.User
		Posts []*db.GetImagePostsByUserDescendingAllRow
	}{
		User:  &db.User{Idusers: cpu.Idusers, Username: user.Username},
		Posts: rows,
	}
	AdminUserImagebbsPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminUserImagebbsPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "ImageBBS", "", &AdminUserProfilePage{}
}

func (p *AdminUserImagebbsPage) PageTitle() string {
	return "User Images"
}

var _ common.Page = (*AdminUserImagebbsPage)(nil)
var _ http.Handler = (*AdminUserImagebbsPage)(nil)

const AdminUserImagebbsPageTmpl tasks.Template = "admin/userImagebbsPage.gohtml"
