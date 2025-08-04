package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminAnnouncementsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Announcements []*db.AdminListAnnouncementsWithNewsRow
		NewsID        string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin Announcements"
	data := Data{CoreData: cd}
	queries := cd.Queries()
	rows, err := queries.AdminListAnnouncementsWithNews(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list announcements: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data.Announcements = rows
	data.NewsID = r.FormValue("news_id")
	handlers.TemplateHandler(w, r, "announcementsPage.gohtml", data)
}
