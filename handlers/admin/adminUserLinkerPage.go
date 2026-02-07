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

type AdminUserLinkerPage struct{}

func (p *AdminUserLinkerPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	rows, err := queries.GetLinkerItemsByUserDescending(r.Context(), db.GetLinkerItemsByUserDescendingParams{
		AuthorID: cpu.Idusers,
		Limit:    100,
		Offset:   0,
	})
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	cd.PageTitle = fmt.Sprintf("Links by %s", user.Username.String)
	data := struct {
		User  *db.User
		Links []*db.GetLinkerItemsByUserDescendingRow
	}{
		User:  &db.User{Idusers: cpu.Idusers, Username: user.Username},
		Links: rows,
	}
	AdminUserLinkerPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminUserLinkerPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Links", "", &AdminUserProfilePage{}
}

func (p *AdminUserLinkerPage) PageTitle() string {
	return "User Links"
}

var _ common.Page = (*AdminUserLinkerPage)(nil)
var _ http.Handler = (*AdminUserLinkerPage)(nil)

const AdminUserLinkerPageTmpl tasks.Template = "admin/userLinkerPage.gohtml"
