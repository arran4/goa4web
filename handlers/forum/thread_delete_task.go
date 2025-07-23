package forum

import (
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// ThreadDeleteTask removes a forum thread.
type ThreadDeleteTask struct{ tasks.TaskString }

var threadDeleteTask = &ThreadDeleteTask{TaskString: TaskForumThreadDelete}

var (
	_ tasks.Task                       = (*ThreadDeleteTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*ThreadDeleteTask)(nil)
)

func (ThreadDeleteTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationForumThreadDeleteEmail")
}

func (ThreadDeleteTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumThreadDeleteEmail")
	return &v
}
