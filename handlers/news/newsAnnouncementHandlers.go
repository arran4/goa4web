package news

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

type AnnouncementAddTask struct{ tasks.TaskString }

var _ tasks.Task = (*AnnouncementAddTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*AnnouncementAddTask)(nil)

func (AnnouncementAddTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationNewsAddEmail")
}

func (AnnouncementAddTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsAddEmail")
	return &v
}

var announcementAddTask = &AnnouncementAddTask{TaskString: TaskAdd}

type AnnouncementDeleteTask struct{ tasks.TaskString }

var announcementDeleteTask = &AnnouncementDeleteTask{TaskString: TaskDelete}

var _ tasks.Task = (*AnnouncementDeleteTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*AnnouncementDeleteTask)(nil)

func (AnnouncementDeleteTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationNewsDeleteEmail")
}

func (AnnouncementDeleteTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsDeleteEmail")
	return &v
}

func (AnnouncementAddTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])

	ann, err := cd.NewsAnnouncement(int32(pid))
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("getLatestAnnouncementByNewsID: %v", err)
		}
	}
	if ann == nil {
		if err := queries.CreateAnnouncement(r.Context(), int32(pid)); err != nil {
			log.Printf("create announcement: %v", err)
		}
	} else if !ann.Active {
		if err := queries.SetAnnouncementActive(r.Context(), db.SetAnnouncementActiveParams{Active: true, ID: ann.ID}); err != nil {
			log.Printf("activate announcement: %v", err)
		}
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
	return nil
}

func (AnnouncementDeleteTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])

	ann, err := cd.NewsAnnouncement(int32(pid))
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("announcementForNews: %v", err)
		}
		handlers.TaskDoneAutoRefreshPage(w, r)
		return nil
	}
	if ann != nil && ann.Active {
		if err := queries.SetAnnouncementActive(r.Context(), db.SetAnnouncementActiveParams{Active: false, ID: ann.ID}); err != nil {
			log.Printf("deactivate announcement: %v", err)
		}
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
	return nil
}
