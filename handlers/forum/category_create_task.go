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
)

func (CategoryCreateTask) AdminEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationForumCategoryCreateEmail")
}

func (CategoryCreateTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumCategoryCreateEmail")
	return &v
}
