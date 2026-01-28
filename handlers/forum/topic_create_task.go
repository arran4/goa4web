package forum

import (
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// TopicCreateTask creates a new forum topic.
type TopicCreateTask struct{ tasks.TaskString }

var topicCreateTask = &TopicCreateTask{TaskString: TaskForumTopicCreate}

var (
	_ tasks.Task                       = (*TopicCreateTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*TopicCreateTask)(nil)
	_ tasks.EmailTemplatesRequired     = (*TopicCreateTask)(nil)
)

func (TopicCreateTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationForumTopicCreate.EmailTemplates(), true
}

func (TopicCreateTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationForumTopicCreate.NotificationTemplate()
	return &v
}

func (TopicCreateTask) RequiredTemplates() []tasks.Template {
	return EmailTemplateAdminNotificationForumTopicCreate.RequiredTemplates()
}
