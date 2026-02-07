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

type AdminUserSubscriptionsPage struct{}

func (p *AdminUserSubscriptionsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	rows, err := queries.ListSubscriptionsByUser(r.Context(), cpu.Idusers)
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	cd.PageTitle = fmt.Sprintf("Subscriptions of %s", user.Username.String)
	data := struct {
		User *db.User
		Subs []*db.ListSubscriptionsByUserRow
	}{
		User: &db.User{Idusers: cpu.Idusers, Username: user.Username},
		Subs: rows,
	}
	AdminUserSubscriptionsPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminUserSubscriptionsPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Subscriptions", "", &AdminUserProfilePage{}
}

func (p *AdminUserSubscriptionsPage) PageTitle() string {
	return "User Subscriptions"
}

var _ common.Page = (*AdminUserSubscriptionsPage)(nil)
var _ http.Handler = (*AdminUserSubscriptionsPage)(nil)

const AdminUserSubscriptionsPageTmpl tasks.Template = "admin/userSubscriptionsPage.gohtml"
