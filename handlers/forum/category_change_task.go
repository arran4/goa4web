package forum

import (
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// CategoryChangeTask updates a forum category name.
type CategoryChangeTask struct{ tasks.TaskString }

var categoryChangeTask = &CategoryChangeTask{TaskString: TaskForumCategoryChange}

const (
	EmailTemplateAdminNotificationForumCategoryChange notif.EmailTemplateName = "adminNotificationForumCategoryChangeEmail"
)

var (
	_ tasks.Task                       = (*CategoryChangeTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*CategoryChangeTask)(nil)
	_ tasks.EmailTemplatesRequired     = (*CategoryChangeTask)(nil)
)

func (CategoryChangeTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationForumCategoryUpdate.EmailTemplates(), true
}

func (CategoryChangeTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationForumCategoryUpdate.NotificationTemplate()
	return &v
}

func (CategoryChangeTask) RequiredTemplates() []tasks.Template {
	return EmailTemplateAdminNotificationForumCategoryUpdate.RequiredTemplates()
}
