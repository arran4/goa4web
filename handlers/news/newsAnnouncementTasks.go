package news

import (
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

var _ tasks.Task = (*AnnouncementAddTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*AnnouncementAddTask)(nil)
var _ tasks.EmailTemplatesRequired = (*AnnouncementAddTask)(nil)

func (AnnouncementAddTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationNewsAdd.EmailTemplates(), true
}

func (AnnouncementAddTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationNewsAdd.NotificationTemplate()
	return &v
}

func (AnnouncementAddTask) EmailTemplatesRequired() []tasks.Page {
	return EmailTemplateAdminNotificationNewsAdd.RequiredPages()
}

var _ tasks.Task = (*AnnouncementDeleteTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*AnnouncementDeleteTask)(nil)
var _ tasks.EmailTemplatesRequired = (*AnnouncementDeleteTask)(nil)

func (AnnouncementDeleteTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationNewsDelete.EmailTemplates(), true
}

func (AnnouncementDeleteTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationNewsDelete.NotificationTemplate()
	return &v
}

func (AnnouncementDeleteTask) EmailTemplatesRequired() []tasks.Page {
	return EmailTemplateAdminNotificationNewsDelete.RequiredPages()
}
