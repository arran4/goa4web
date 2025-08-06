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

func (RemakeWritingFinishedTask) Action(http.ResponseWriter, *http.Request) any { return nil }

func (RemakeWritingFinishedTask) AdminEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return notif.NewEmailTemplates("searchRebuildWritingEmail")
}

func (RemakeWritingFinishedTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := notif.NotificationTemplateFilenameGenerator("search_rebuild_writing")
	return &s
}

func (RemakeWritingFinishedTask) SelfEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return notif.NewEmailTemplates("searchRebuildWritingEmail")
}

func (RemakeWritingFinishedTask) SelfInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := notif.NotificationTemplateFilenameGenerator("search_rebuild_writing")
	return &s
}
