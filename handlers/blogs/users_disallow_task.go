package blogs

import (
	"net/http"

	"github.com/arran4/goa4web/internal/eventbus"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// UsersDisallowTask removes permissions from multiple users.
type UsersDisallowTask struct{ tasks.TaskString }

var usersDisallowTask = &UsersDisallowTask{TaskString: TaskUsersDisallow}

var _ tasks.Task = (*UsersDisallowTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*UsersDisallowTask)(nil)

func (UsersDisallowTask) AdminEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationBlogUsersDisallowEmail")
}

func (UsersDisallowTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationBlogUsersDisallowEmail")
	return &v
}

func (UsersDisallowTask) Action(w http.ResponseWriter, r *http.Request) any {
	UsersPermissionsBulkDisallowPage(w, r)
	return nil
}
