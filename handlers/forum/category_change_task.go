package forum

import (
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// CategoryChangeTask updates a forum category name.
type CategoryChangeTask struct{ tasks.TaskString }

var categoryChangeTask = &CategoryChangeTask{TaskString: TaskForumCategoryChange}

var (
	_ tasks.Task                       = (*CategoryChangeTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*CategoryChangeTask)(nil)
)

func (CategoryChangeTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationForumCategoryChangeEmail")
}

func (CategoryChangeTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumCategoryChangeEmail")
	return &v
}
