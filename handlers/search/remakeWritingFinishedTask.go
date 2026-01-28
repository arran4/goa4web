package search

import (
	"net/http"

	"github.com/arran4/goa4web/internal/eventbus"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeWritingFinishedTask notifies when the writing index rebuild completes.
type RemakeWritingFinishedTask struct{ tasks.TaskString }

var remakeWritingFinishedTask = &RemakeWritingFinishedTask{TaskString: TaskRemakeWritingSearchComplete}

var _ tasks.Task = (*RemakeWritingFinishedTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*RemakeWritingFinishedTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*RemakeWritingFinishedTask)(nil)
var _ tasks.EmailTemplatesRequired = (*RemakeWritingFinishedTask)(nil)

func (RemakeWritingFinishedTask) Action(http.ResponseWriter, *http.Request) any { return nil }

func (RemakeWritingFinishedTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateSearchRebuildWriting.EmailTemplates(), true
}

func (RemakeWritingFinishedTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateSearchRebuildWriting.NotificationTemplate()
	return &s
}

func (RemakeWritingFinishedTask) SelfEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateSearchRebuildWriting.EmailTemplates(), true
}

func (RemakeWritingFinishedTask) SelfInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateSearchRebuildWriting.NotificationTemplate()
	return &s
}

func (RemakeWritingFinishedTask) RequiredTemplates() []tasks.Template {
	return EmailTemplateSearchRebuildWriting.RequiredTemplates()
}
