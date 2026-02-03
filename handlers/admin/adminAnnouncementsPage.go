package admin

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminAnnouncementsPage struct{}

func (p *AdminAnnouncementsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Announcements []*db.AdminListAnnouncementsWithNewsRow
		NewsID        string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin Announcements"
	data := Data{}
	queries := cd.Queries()
	rows, err := queries.AdminListAnnouncementsWithNews(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	data.Announcements = rows
	data.NewsID = r.FormValue("news_id")
	AdminAnnouncementsPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminAnnouncementsPage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Announcements", "/admin/announcements", &AdminPage{}
}

func (p *AdminAnnouncementsPage) PageTitle() string {
	return "Admin Announcements"
}

var _ tasks.Page = (*AdminAnnouncementsPage)(nil)
var _ http.Handler = (*AdminAnnouncementsPage)(nil)

const AdminAnnouncementsPageTmpl tasks.Template = "admin/announcementsPage.gohtml"
