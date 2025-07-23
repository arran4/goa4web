package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// AddAnnouncementTask posts a new announcement.
type AddAnnouncementTask struct{ tasks.TaskString }

var addAnnouncementTask = &AddAnnouncementTask{TaskString: TaskAdd}

// DeleteAnnouncementTask removes an announcement.
type DeleteAnnouncementTask struct{ tasks.TaskString }

var deleteAnnouncementTask = &DeleteAnnouncementTask{TaskString: TaskDelete}

var _ tasks.Task = (*AddAnnouncementTask)(nil)
var _ tasks.AuditableTask = (*AddAnnouncementTask)(nil)

// addAnnouncementTask notifies admins so they know announcements were updated.
var _ notif.AdminEmailTemplateProvider = (*AddAnnouncementTask)(nil)

var _ tasks.Task = (*DeleteAnnouncementTask)(nil)
var _ tasks.AuditableTask = (*DeleteAnnouncementTask)(nil)

// deleteAnnouncementTask also notifies admins of changes for transparency.
var _ notif.AdminEmailTemplateProvider = (*DeleteAnnouncementTask)(nil)

func (AddAnnouncementTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	nid, err := strconv.Atoi(r.PostFormValue("news_id"))
	if err != nil {
		return fmt.Errorf("news id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := queries.CreateAnnouncement(r.Context(), int32(nid)); err != nil {
		return fmt.Errorf("create announcement fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["NewsID"] = nid
		}
	}
	return nil
}

func (AddAnnouncementTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("announcementEmail")
}

func (AddAnnouncementTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("announcement")
	return &v
}

func (DeleteAnnouncementTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		if err := queries.DeleteAnnouncement(r.Context(), int32(id)); err != nil {
			return fmt.Errorf("delete announcement fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["AnnouncementID"] = id
			}
		}
	}
	return nil
}

func (DeleteAnnouncementTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("announcementEmail")
}

func (DeleteAnnouncementTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("announcement")
	return &v
}

// AuditRecord summarises an announcement being created.
func (AddAnnouncementTask) AuditRecord(data map[string]any) string {
	if id, ok := data["NewsID"].(int); ok {
		return fmt.Sprintf("announcement created for news %d", id)
	}
	return "announcement created"
}

// AuditRecord summarises an announcement deletion.
func (DeleteAnnouncementTask) AuditRecord(data map[string]any) string {
	if id, ok := data["AnnouncementID"].(int); ok {
		return fmt.Sprintf("announcement %d deleted", id)
	}
	return "announcement deleted"
}
