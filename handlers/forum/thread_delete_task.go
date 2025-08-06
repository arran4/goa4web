package forum

import (
	"github.com/arran4/goa4web/internal/eventbus"
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

func (ThreadDeleteTask) AdminEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationForumThreadDeleteEmail")
}

func (ThreadDeleteTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumThreadDeleteEmail")
	return &v
}
