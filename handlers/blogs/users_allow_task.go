package blogs

import (
	"net/http"

	"github.com/arran4/goa4web/internal/eventbus"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// UsersAllowTask grants multiple users permissions.
type UsersAllowTask struct{ tasks.TaskString }

var usersAllowTask = &UsersAllowTask{TaskString: TaskUsersAllow}

var _ tasks.Task = (*UsersAllowTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*UsersAllowTask)(nil)

func (UsersAllowTask) AdminEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationBlogUsersAllowEmail")
}

func (UsersAllowTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationBlogUsersAllowEmail")
	return &v
}

func (UsersAllowTask) Action(w http.ResponseWriter, r *http.Request) any {
	UsersPermissionsBulkAllowPage(w, r)
	return nil
}
