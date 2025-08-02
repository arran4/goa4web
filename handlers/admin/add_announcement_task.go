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

var _ tasks.Task = (*AddAnnouncementTask)(nil)
var _ tasks.AuditableTask = (*AddAnnouncementTask)(nil)

// addAnnouncementTask notifies admins so they know announcements were updated.
var _ notif.AdminEmailTemplateProvider = (*AddAnnouncementTask)(nil)

func (AddAnnouncementTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	nid, err := strconv.Atoi(r.PostFormValue("news_id"))
	if err != nil {
		return fmt.Errorf("news id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := queries.AdminPromoteAnnouncement(r.Context(), int32(nid)); err != nil {
		return fmt.Errorf("promote announcement fail %w", handlers.ErrRedirectOnSamePageHandler(err))
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

// AuditRecord summarises an announcement being created.
func (AddAnnouncementTask) AuditRecord(data map[string]any) string {
	if id, ok := data["NewsID"].(int); ok {
		return fmt.Sprintf("announcement created for news %d", id)
	}
	return "announcement created"
}
