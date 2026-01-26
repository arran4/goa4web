package search

import (
	"net/http"

	"github.com/arran4/goa4web/internal/eventbus"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeBlogFinishedTask notifies when the blog index rebuild completes.
type RemakeBlogFinishedTask struct{ tasks.TaskString }

var remakeBlogFinishedTask = &RemakeBlogFinishedTask{TaskString: TaskRemakeBlogSearchComplete}

var _ tasks.Task = (*RemakeBlogFinishedTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*RemakeBlogFinishedTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*RemakeBlogFinishedTask)(nil)
var _ tasks.EmailTemplatesRequired = (*RemakeBlogFinishedTask)(nil)

func (RemakeBlogFinishedTask) Action(http.ResponseWriter, *http.Request) any { return nil }

func (RemakeBlogFinishedTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateSearchRebuildBlog.EmailTemplates(), true
}

func (RemakeBlogFinishedTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateSearchRebuildBlog.NotificationTemplate()
	return &s
}

func (RemakeBlogFinishedTask) SelfEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateSearchRebuildBlog.EmailTemplates(), true
}

func (RemakeBlogFinishedTask) SelfInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateSearchRebuildBlog.NotificationTemplate()
	return &s
}

func (RemakeBlogFinishedTask) EmailTemplatesRequired() []tasks.Page {
	return EmailTemplateSearchRebuildBlog.RequiredPages()
}
