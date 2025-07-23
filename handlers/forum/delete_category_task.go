package forum

import (
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// DeleteCategoryTask removes a forum category.
type DeleteCategoryTask struct{ tasks.TaskString }

var deleteCategoryTask = &DeleteCategoryTask{TaskString: TaskDeleteCategory}

var (
	_ tasks.Task                       = (*DeleteCategoryTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*DeleteCategoryTask)(nil)
)

func (DeleteCategoryTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationForumDeleteCategoryEmail")
}

func (DeleteCategoryTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumDeleteCategoryEmail")
	return &v
}
