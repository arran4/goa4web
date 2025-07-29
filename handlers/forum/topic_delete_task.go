package forum

import (
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// TopicDeleteTask removes a forum topic.
type TopicDeleteTask struct{ tasks.TaskString }

var topicDeleteTask = &TopicDeleteTask{TaskString: TaskForumTopicDelete}

var (
	_ tasks.Task                       = (*TopicDeleteTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*TopicDeleteTask)(nil)
)

func (TopicDeleteTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationForumTopicDeleteEmail")
}

func (TopicDeleteTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumTopicDeleteEmail")
	return &v
}
