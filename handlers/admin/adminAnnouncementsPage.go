package admin

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/core/common"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

type addAnnouncementTask struct{ tasks.TaskString }
type deleteAnnouncementTask struct{ tasks.TaskString }

var _ tasks.Task = (*addAnnouncementTask)(nil)

// addAnnouncementTask notifies admins so they know announcements were updated.
var _ notif.AdminEmailTemplateProvider = (*addAnnouncementTask)(nil)

var _ tasks.Task = (*deleteAnnouncementTask)(nil)

// deleteAnnouncementTask also notifies admins of changes for transparency.
var _ notif.AdminEmailTemplateProvider = (*deleteAnnouncementTask)(nil)

func AdminAnnouncementsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Announcements []*db.ListAnnouncementsWithNewsRow
		NewsID        string
	}
	data := Data{CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData)}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	rows, err := queries.ListAnnouncementsWithNews(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list announcements: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Announcements = rows
	data.NewsID = r.FormValue("news_id")
	handlers.TemplateHandler(w, r, "announcementsPage.gohtml", data)
}

func (addAnnouncementTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	nid, err := strconv.Atoi(r.PostFormValue("news_id"))
	if err != nil {
		log.Printf("news id: %v", err)
		handlers.TaskDoneAutoRefreshPage(w, r)
		return
	}
	if err := queries.CreateAnnouncement(r.Context(), int32(nid)); err != nil {
		log.Printf("create announcement: %v", err)
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}

func (addAnnouncementTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("announcementEmail")
}

func (addAnnouncementTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("announcement")
	return &v
}

func (deleteAnnouncementTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm: %v", err)
	}
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		if err := queries.DeleteAnnouncement(r.Context(), int32(id)); err != nil {
			log.Printf("delete announcement: %v", err)
		}
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}

func (deleteAnnouncementTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("announcementEmail")
}

func (deleteAnnouncementTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("announcement")
	return &v
}
