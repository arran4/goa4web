package forum

import (
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// TopicChangeTask updates a forum topic title or details.
type TopicChangeTask struct{ tasks.TaskString }

var topicChangeTask = &TopicChangeTask{TaskString: TaskForumTopicChange}

var (
	_ tasks.Task                       = (*TopicChangeTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*TopicChangeTask)(nil)
)

func (TopicChangeTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationForumTopicChangeEmail")
}

func (TopicChangeTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumTopicChangeEmail")
	return &v
}
