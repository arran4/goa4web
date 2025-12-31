package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// DeleteAnnouncementTask removes an announcement.
type DeleteAnnouncementTask struct{ tasks.TaskString }

var deleteAnnouncementTask = &DeleteAnnouncementTask{TaskString: TaskDelete}

var _ tasks.Task = (*DeleteAnnouncementTask)(nil)
var _ tasks.AuditableTask = (*DeleteAnnouncementTask)(nil)

// deleteAnnouncementTask also notifies admins of changes for transparency.
var _ notif.AdminEmailTemplateProvider = (*DeleteAnnouncementTask)(nil)

func (DeleteAnnouncementTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasRole("administrator") {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { handlers.RenderErrorPage(w, r, handlers.ErrForbidden) })
	}
	queries := cd.Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		if err := queries.AdminDemoteAnnouncement(r.Context(), int32(id)); err != nil {
			return fmt.Errorf("demote announcement fail %w", handlers.ErrRedirectOnSamePageHandler(err))
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

func (DeleteAnnouncementTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("announcementEmail"), true
}

func (DeleteAnnouncementTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("announcement")
	return &v
}

// AuditRecord summarises an announcement deletion.
func (DeleteAnnouncementTask) AuditRecord(data map[string]any) string {
	if id, ok := data["AnnouncementID"].(int); ok {
		return fmt.Sprintf("announcement %d deleted", id)
	}
	return "announcement deleted"
}
