package forum

import (
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// TopicChangeTask updates a forum topic title or details.
type TopicChangeTask struct{ tasks.TaskString }

var topicChangeTask = &TopicChangeTask{TaskString: TaskForumTopicChange}

var (
	_ tasks.Task                       = (*TopicChangeTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*TopicChangeTask)(nil)
	_ tasks.EmailTemplatesRequired     = (*TopicChangeTask)(nil)
)

func (TopicChangeTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationForumTopicChange.EmailTemplates(), true
}

func (TopicChangeTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationForumTopicChange.NotificationTemplate()
	return &v
}

func (TopicChangeTask) RequiredTemplates() []tasks.Template {
	return EmailTemplateAdminNotificationForumTopicChange.RequiredTemplates()
}
