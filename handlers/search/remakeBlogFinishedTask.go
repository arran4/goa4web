package search

import (
	"net/http"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeBlogFinishedTask notifies when the blog index rebuild completes.
type RemakeBlogFinishedTask struct{ tasks.TaskString }

var remakeBlogFinishedTask = &RemakeBlogFinishedTask{TaskString: TaskRemakeBlogSearchComplete}

var _ tasks.Task = (*RemakeBlogFinishedTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*RemakeBlogFinishedTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*RemakeBlogFinishedTask)(nil)

func (RemakeBlogFinishedTask) Action(http.ResponseWriter, *http.Request) any { return nil }

func (RemakeBlogFinishedTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("searchRebuildBlogEmail")
}

func (RemakeBlogFinishedTask) AdminInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("search_rebuild_blog")
	return &s
}

func (RemakeBlogFinishedTask) SelfEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("searchRebuildBlogEmail")
}

func (RemakeBlogFinishedTask) SelfInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("search_rebuild_blog")
	return &s
}
