package blogs

import (
	"net/http"

	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// UsersDisallowTask removes permissions from multiple users.
type UsersDisallowTask struct{ tasks.TaskString }

var usersDisallowTask = &UsersDisallowTask{TaskString: TaskUsersDisallow}

var _ tasks.Task = (*UsersDisallowTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*UsersDisallowTask)(nil)

func (UsersDisallowTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationBlogUsersDisallowEmail")
}

func (UsersDisallowTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationBlogUsersDisallowEmail")
	return &v
}

func (UsersDisallowTask) Action(w http.ResponseWriter, r *http.Request) any {
	UsersPermissionsBulkDisallowPage(w, r)
	return nil
}
