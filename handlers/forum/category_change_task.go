package forum

import (
	"github.com/arran4/goa4web/handlers/forumcommon"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// CategoryChangeTask updates a forum category name.
type CategoryChangeTask struct{ tasks.TaskString }

var categoryChangeTask = &CategoryChangeTask{TaskString: forumcommon.TaskForumCategoryChange}

var (
	_ tasks.Task                       = (*CategoryChangeTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*CategoryChangeTask)(nil)
)

func (CategoryChangeTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("adminNotificationForumCategoryChangeEmail"), true
}

func (CategoryChangeTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumCategoryChangeEmail")
	return &v
}
