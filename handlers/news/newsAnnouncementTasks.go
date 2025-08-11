package news

import (
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

var _ tasks.Task = (*AnnouncementAddTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*AnnouncementAddTask)(nil)

func (AnnouncementAddTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("adminNotificationNewsAddEmail"), true
}

func (AnnouncementAddTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsAddEmail")
	return &v
}

var _ tasks.Task = (*AnnouncementDeleteTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*AnnouncementDeleteTask)(nil)

func (AnnouncementDeleteTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("adminNotificationNewsDeleteEmail"), true
}

func (AnnouncementDeleteTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsDeleteEmail")
	return &v
}
