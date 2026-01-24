package forum

import (
	"github.com/arran4/goa4web/handlers/forumcommon"
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
)

func (TopicDeleteTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("adminNotificationForumTopicDeleteEmail"), true
}

func (TopicDeleteTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumTopicDeleteEmail")
	return &v
}
