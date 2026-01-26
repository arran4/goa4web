package search

import (
	"net/http"

	"github.com/arran4/goa4web/internal/eventbus"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeLinkerFinishedTask notifies when the linker index rebuild completes.
type RemakeLinkerFinishedTask struct{ tasks.TaskString }

var remakeLinkerFinishedTask = &RemakeLinkerFinishedTask{TaskString: TaskRemakeLinkerSearchComplete}

var _ tasks.Task = (*RemakeLinkerFinishedTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*RemakeLinkerFinishedTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*RemakeLinkerFinishedTask)(nil)
var _ tasks.EmailTemplatesRequired = (*RemakeLinkerFinishedTask)(nil)

func (RemakeLinkerFinishedTask) Action(http.ResponseWriter, *http.Request) any { return nil }

func (RemakeLinkerFinishedTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateSearchRebuildLinker.EmailTemplates(), true
}

func (RemakeLinkerFinishedTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateSearchRebuildLinker.NotificationTemplate()
	return &s
}

func (RemakeLinkerFinishedTask) SelfEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateSearchRebuildLinker.EmailTemplates(), true
}

func (RemakeLinkerFinishedTask) SelfInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateSearchRebuildLinker.NotificationTemplate()
	return &s
}

func (RemakeLinkerFinishedTask) RequiredTemplates() []tasks.Template {
	return EmailTemplateSearchRebuildLinker.RequiredTemplates()
}
