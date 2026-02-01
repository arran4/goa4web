package admin

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminAnnouncementsPageTask struct{}

func (t *AdminAnnouncementsPageTask) Action(w http.ResponseWriter, r *http.Request) any {
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
		return err
	}
	data.Announcements = rows
	data.NewsID = r.FormValue("news_id")
	return AdminAnnouncementsPageTmpl.Handler(data)
}

func (t *AdminAnnouncementsPageTask) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Announcements", "/admin/announcements", &AdminPageTask{}
}

var _ tasks.Task = (*AdminAnnouncementsPageTask)(nil)
var _ tasks.HasBreadcrumb = (*AdminAnnouncementsPageTask)(nil)

const AdminAnnouncementsPageTmpl tasks.Template = "admin/announcementsPage.gohtml"
