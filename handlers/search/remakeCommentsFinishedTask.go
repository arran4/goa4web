package search

import (
	"net/http"

	"github.com/arran4/goa4web/internal/eventbus"

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

func (RemakeCommentsFinishedTask) AdminEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return notif.NewEmailTemplates("searchRebuildCommentsEmail")
}

func (RemakeCommentsFinishedTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := notif.NotificationTemplateFilenameGenerator("search_rebuild_comments")
	return &s
}

func (RemakeCommentsFinishedTask) SelfEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return notif.NewEmailTemplates("searchRebuildCommentsEmail")
}

func (RemakeCommentsFinishedTask) SelfInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := notif.NotificationTemplateFilenameGenerator("search_rebuild_comments")
	return &s
}
