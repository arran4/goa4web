package forum

import (
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// DeleteCategoryTask removes a forum category.
type DeleteCategoryTask struct{ tasks.TaskString }

var deleteCategoryTask = &DeleteCategoryTask{TaskString: TaskDeleteCategory}

const (
	EmailTemplateAdminNotificationForumDeleteCategory notif.EmailTemplateName = "adminNotificationForumDeleteCategoryEmail"
)

var (
	_ tasks.Task                       = (*DeleteCategoryTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*DeleteCategoryTask)(nil)
	_ tasks.EmailTemplatesRequired     = (*DeleteCategoryTask)(nil)
)

func (DeleteCategoryTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationForumCategoryDelete.EmailTemplates(), true
}

func (DeleteCategoryTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationForumCategoryDelete.NotificationTemplate()
	return &v
}

func (DeleteCategoryTask) RequiredTemplates() []tasks.Template {
	return EmailTemplateAdminNotificationForumCategoryDelete.RequiredTemplates()
}
