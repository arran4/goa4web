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

func (RemakeNewsFinishedTask) Action(http.ResponseWriter, *http.Request) any { return nil }

func (RemakeNewsFinishedTask) AdminEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return notif.NewEmailTemplates("searchRebuildNewsEmail")
}

func (RemakeNewsFinishedTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := notif.NotificationTemplateFilenameGenerator("search_rebuild_news")
	return &s
}

func (RemakeNewsFinishedTask) SelfEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return notif.NewEmailTemplates("searchRebuildNewsEmail")
}

func (RemakeNewsFinishedTask) SelfInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := notif.NotificationTemplateFilenameGenerator("search_rebuild_news")
	return &s
}
