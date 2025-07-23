package blogs

import (
	"net/http"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// UsersAllowTask grants multiple users permissions.
type UsersAllowTask struct{ tasks.TaskString }

var usersAllowTask = &UsersAllowTask{TaskString: TaskUsersAllow}

var _ tasks.Task = (*UsersAllowTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*UsersAllowTask)(nil)

func (UsersAllowTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationBlogUsersAllowEmail")
}

func (UsersAllowTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationBlogUsersAllowEmail")
	return &v
}

func (UsersAllowTask) Action(w http.ResponseWriter, r *http.Request) any {
	UsersPermissionsBulkAllowPage(w, r)
	return nil
}
