package forum

import (
	"github.com/arran4/goa4web/handlers/forumcommon"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// DeleteCategoryTask removes a forum category.
type DeleteCategoryTask struct{ tasks.TaskString }

var deleteCategoryTask = &DeleteCategoryTask{TaskString: forumcommon.TaskDeleteCategory}

var (
	_ tasks.Task                       = (*DeleteCategoryTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*DeleteCategoryTask)(nil)
)

func (DeleteCategoryTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("adminNotificationForumDeleteCategoryEmail"), true
}

func (DeleteCategoryTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumDeleteCategoryEmail")
	return &v
}
