package forum

import (
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// CategoryCreateTask creates a new forum category.
type CategoryCreateTask struct{ tasks.TaskString }

var categoryCreateTask = &CategoryCreateTask{TaskString: TaskForumCategoryCreate}

var (
	_ tasks.Task                       = (*CategoryCreateTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*CategoryCreateTask)(nil)
	_ tasks.EmailTemplatesRequired     = (*CategoryCreateTask)(nil)
)

func (CategoryCreateTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationForumCategoryCreate.EmailTemplates(), true
}

func (CategoryCreateTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationForumCategoryCreate.NotificationTemplate()
	return &v
}

func (CategoryCreateTask) EmailTemplatesRequired() []tasks.Page {
	return EmailTemplateAdminNotificationForumCategoryCreate.RequiredPages()
}
