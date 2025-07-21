package admin

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AddAnnouncementTask posts a new announcement.
type AddAnnouncementTask struct{ tasks.TaskString }

var addAnnouncementTask = &AddAnnouncementTask{TaskString: TaskAdd}

// DeleteAnnouncementTask removes an announcement.
type DeleteAnnouncementTask struct{ tasks.TaskString }

var deleteAnnouncementTask = &DeleteAnnouncementTask{TaskString: TaskDelete}

var _ tasks.Task = (*AddAnnouncementTask)(nil)

// addAnnouncementTask notifies admins so they know announcements were updated.
var _ notif.AdminEmailTemplateProvider = (*AddAnnouncementTask)(nil)

var _ tasks.Task = (*DeleteAnnouncementTask)(nil)

// deleteAnnouncementTask also notifies admins of changes for transparency.
var _ notif.AdminEmailTemplateProvider = (*DeleteAnnouncementTask)(nil)

func AdminAnnouncementsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Announcements []*db.ListAnnouncementsWithNewsRow
		NewsID        string
	}
	data := Data{CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData)}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
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

func (AddAnnouncementTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
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

func (AddAnnouncementTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("announcementEmail")
}

func (AddAnnouncementTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("announcement")
	return &v
}

func (DeleteAnnouncementTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
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

func (DeleteAnnouncementTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("announcementEmail")
}

func (DeleteAnnouncementTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("announcement")
	return &v
}
