package search

import (
	"net/http"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeImageFinishedTask notifies when the image index rebuild completes.
type RemakeImageFinishedTask struct{ tasks.TaskString }

var remakeImageFinishedTask = &RemakeImageFinishedTask{TaskString: TaskRemakeImageSearchComplete}

var _ tasks.Task = (*RemakeImageFinishedTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*RemakeImageFinishedTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*RemakeImageFinishedTask)(nil)

func (RemakeImageFinishedTask) Action(http.ResponseWriter, *http.Request) any { return nil }

func (RemakeImageFinishedTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("searchRebuildImageEmail")
}

func (RemakeImageFinishedTask) AdminInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("search_rebuild_image")
	return &s
}

func (RemakeImageFinishedTask) SelfEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("searchRebuildImageEmail")
}

func (RemakeImageFinishedTask) SelfInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("search_rebuild_image")
	return &s
}
