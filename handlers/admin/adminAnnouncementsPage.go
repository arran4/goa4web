package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminAnnouncementsPage(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("list announcements: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data.Announcements = rows
	data.NewsID = r.FormValue("news_id")
	AdminAnnouncementsPageTmpl.Handle(w, r, data)
}

const AdminAnnouncementsPageTmpl tasks.Template = "admin/announcementsPage.gohtml"
