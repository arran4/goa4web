package forum

import (
	"github.com/arran4/goa4web/handlers/forumcommon"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// TopicCreateTask creates a new forum topic.
type TopicCreateTask struct{ tasks.TaskString }

var topicCreateTask = &TopicCreateTask{TaskString: forumcommon.TaskForumTopicCreate}

var (
	_ tasks.Task                       = (*TopicCreateTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*TopicCreateTask)(nil)
)

func (TopicCreateTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("adminNotificationForumTopicCreateEmail"), true
}

func (TopicCreateTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumTopicCreateEmail")
	return &v
}
