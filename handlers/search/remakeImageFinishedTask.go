package search

import (
	"net/http"

	"github.com/arran4/goa4web/internal/eventbus"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeImageFinishedTask notifies when the image index rebuild completes.
type RemakeImageFinishedTask struct{ tasks.TaskString }

var remakeImageFinishedTask = &RemakeImageFinishedTask{TaskString: TaskRemakeImageSearchComplete}

var _ tasks.Task = (*RemakeImageFinishedTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*RemakeImageFinishedTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*RemakeImageFinishedTask)(nil)
var _ tasks.EmailTemplatesRequired = (*RemakeImageFinishedTask)(nil)

func (RemakeImageFinishedTask) Action(http.ResponseWriter, *http.Request) any { return nil }

func (RemakeImageFinishedTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateSearchRebuildImage.EmailTemplates(), true
}

func (RemakeImageFinishedTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateSearchRebuildImage.NotificationTemplate()
	return &s
}

func (RemakeImageFinishedTask) SelfEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateSearchRebuildImage.EmailTemplates(), true
}

func (RemakeImageFinishedTask) SelfInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateSearchRebuildImage.NotificationTemplate()
	return &s
}

func (RemakeImageFinishedTask) EmailTemplatesRequired() []tasks.Page {
	return EmailTemplateSearchRebuildImage.RequiredPages()
}
