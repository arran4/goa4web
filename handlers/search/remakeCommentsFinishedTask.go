package search

import (
	"net/http"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeCommentsFinishedTask notifies when the comment index rebuild completes.
type RemakeCommentsFinishedTask struct{ tasks.TaskString }

var remakeCommentsFinishedTask = &RemakeCommentsFinishedTask{TaskString: TaskRemakeCommentsSearchComplete}

var _ tasks.Task = (*RemakeCommentsFinishedTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*RemakeCommentsFinishedTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*RemakeCommentsFinishedTask)(nil)

func (RemakeCommentsFinishedTask) Action(http.ResponseWriter, *http.Request) any { return nil }

func (RemakeCommentsFinishedTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("searchRebuildCommentsEmail")
}

func (RemakeCommentsFinishedTask) AdminInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("search_rebuild_comments")
	return &s
}

func (RemakeCommentsFinishedTask) SelfEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("searchRebuildCommentsEmail")
}

func (RemakeCommentsFinishedTask) SelfInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("search_rebuild_comments")
	return &s
}
