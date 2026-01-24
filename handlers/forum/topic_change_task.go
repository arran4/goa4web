package forum

import (
	"github.com/arran4/goa4web/handlers/forumcommon"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// TopicChangeTask updates a forum topic title or details.
type TopicChangeTask struct{ tasks.TaskString }

var topicChangeTask = &TopicChangeTask{TaskString: forumcommon.TaskForumTopicChange}

var (
	_ tasks.Task                       = (*TopicChangeTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*TopicChangeTask)(nil)
)

func (TopicChangeTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("adminNotificationForumTopicChangeEmail"), true
}

func (TopicChangeTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumTopicChangeEmail")
	return &v
}
