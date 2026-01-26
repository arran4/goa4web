package search

import (
	"net/http"

	"github.com/arran4/goa4web/internal/eventbus"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeNewsFinishedTask notifies when the news index rebuild completes.
type RemakeNewsFinishedTask struct{ tasks.TaskString }

var remakeNewsFinishedTask = &RemakeNewsFinishedTask{TaskString: TaskRemakeNewsSearchComplete}

var _ tasks.Task = (*RemakeNewsFinishedTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*RemakeNewsFinishedTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*RemakeNewsFinishedTask)(nil)
var _ tasks.EmailTemplatesRequired = (*RemakeNewsFinishedTask)(nil)

func (RemakeNewsFinishedTask) Action(http.ResponseWriter, *http.Request) any { return nil }

func (RemakeNewsFinishedTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateSearchRebuildNews.EmailTemplates(), true
}

func (RemakeNewsFinishedTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateSearchRebuildNews.NotificationTemplate()
	return &s
}

func (RemakeNewsFinishedTask) SelfEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateSearchRebuildNews.EmailTemplates(), true
}

func (RemakeNewsFinishedTask) SelfInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateSearchRebuildNews.NotificationTemplate()
	return &s
}

func (RemakeNewsFinishedTask) EmailTemplatesRequired() []tasks.Page {
	return EmailTemplateSearchRebuildNews.RequiredPages()
}
