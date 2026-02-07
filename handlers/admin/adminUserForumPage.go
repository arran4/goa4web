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

type AdminUserForumPage struct{}

func (p *AdminUserForumPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	rows, err := queries.AdminGetThreadsStartedByUserWithTopic(r.Context(), cpu.Idusers)
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	cd.PageTitle = fmt.Sprintf("Forum threads by %s", user.Username.String)
	data := struct {
		User    *db.User
		Threads []*db.AdminGetThreadsStartedByUserWithTopicRow
	}{
		User:    &db.User{Idusers: cpu.Idusers, Username: user.Username},
		Threads: rows,
	}
	AdminUserForumPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminUserForumPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Forum Threads", "", &AdminUserProfilePage{}
}

func (p *AdminUserForumPage) PageTitle() string {
	return "User Forum Threads"
}

var _ common.Page = (*AdminUserForumPage)(nil)
var _ http.Handler = (*AdminUserForumPage)(nil)

const AdminUserForumPageTmpl tasks.Template = "admin/userForumPage.gohtml"
