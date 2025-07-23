package news

import (
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

var _ tasks.Task = (*AnnouncementAddTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*AnnouncementAddTask)(nil)

func (AnnouncementAddTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationNewsAddEmail")
}

func (AnnouncementAddTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsAddEmail")
	return &v
}

var _ tasks.Task = (*AnnouncementDeleteTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*AnnouncementDeleteTask)(nil)

func (AnnouncementDeleteTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationNewsDeleteEmail")
}

func (AnnouncementDeleteTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsDeleteEmail")
	return &v
}
