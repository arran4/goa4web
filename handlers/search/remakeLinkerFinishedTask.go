package search

import (
	"net/http"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeLinkerFinishedTask notifies when the linker index rebuild completes.
type RemakeLinkerFinishedTask struct{ tasks.TaskString }

var remakeLinkerFinishedTask = &RemakeLinkerFinishedTask{TaskString: TaskRemakeLinkerSearchComplete}

var _ tasks.Task = (*RemakeLinkerFinishedTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*RemakeLinkerFinishedTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*RemakeLinkerFinishedTask)(nil)

func (RemakeLinkerFinishedTask) Action(http.ResponseWriter, *http.Request) any { return nil }

func (RemakeLinkerFinishedTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("searchRebuildLinkerEmail")
}

func (RemakeLinkerFinishedTask) AdminInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("search_rebuild_linker")
	return &s
}

func (RemakeLinkerFinishedTask) SelfEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("searchRebuildLinkerEmail")
}

func (RemakeLinkerFinishedTask) SelfInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("search_rebuild_linker")
	return &s
}
