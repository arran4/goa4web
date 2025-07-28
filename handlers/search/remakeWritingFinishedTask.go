package search

import (
	"net/http"

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

func (RemakeWritingFinishedTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("searchRebuildWritingEmail")
}

func (RemakeWritingFinishedTask) AdminInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("search_rebuild_writing")
	return &s
}

func (RemakeWritingFinishedTask) SelfEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("searchRebuildWritingEmail")
}

func (RemakeWritingFinishedTask) SelfInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("search_rebuild_writing")
	return &s
}
