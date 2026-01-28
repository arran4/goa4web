package forum

import (
	"github.com/arran4/goa4web/handlers/forum/forumcommon"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// TopicDeleteTask removes a forum topic.
type TopicDeleteTask struct{ tasks.TaskString }

var topicDeleteTask = &TopicDeleteTask{TaskString: forumcommon.TaskForumTopicDelete}

var (
	_ tasks.Task                       = (*TopicDeleteTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*TopicDeleteTask)(nil)
	_ tasks.EmailTemplatesRequired     = (*TopicDeleteTask)(nil)
)

func (TopicDeleteTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationForumTopicDelete.EmailTemplates(), true
}

func (TopicDeleteTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationForumTopicDelete.NotificationTemplate()
	return &v
}

func (TopicDeleteTask) RequiredTemplates() []tasks.Template {
	return EmailTemplateAdminNotificationForumTopicDelete.RequiredTemplates()
}
