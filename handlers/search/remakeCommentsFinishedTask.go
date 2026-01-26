package search

import (
	"net/http"

	"github.com/arran4/goa4web/internal/eventbus"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeCommentsFinishedTask notifies when the comment index rebuild completes.
type RemakeCommentsFinishedTask struct{ tasks.TaskString }

var remakeCommentsFinishedTask = &RemakeCommentsFinishedTask{TaskString: TaskRemakeCommentsSearchComplete}

var _ tasks.Task = (*RemakeCommentsFinishedTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*RemakeCommentsFinishedTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*RemakeCommentsFinishedTask)(nil)
var _ tasks.EmailTemplatesRequired = (*RemakeCommentsFinishedTask)(nil)

func (RemakeCommentsFinishedTask) Action(http.ResponseWriter, *http.Request) any { return nil }

func (RemakeCommentsFinishedTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateSearchRebuildComments.EmailTemplates(), true
}

func (RemakeCommentsFinishedTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateSearchRebuildComments.NotificationTemplate()
	return &s
}

func (RemakeCommentsFinishedTask) SelfEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateSearchRebuildComments.EmailTemplates(), true
}

func (RemakeCommentsFinishedTask) SelfInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateSearchRebuildComments.NotificationTemplate()
	return &s
}

func (RemakeCommentsFinishedTask) RequiredTemplates() []tasks.Template {
	return EmailTemplateSearchRebuildComments.RequiredTemplates()
}
