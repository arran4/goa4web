package blogs

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// UserDisallowTask removes a user's permission.
type UserDisallowTask struct{ tasks.TaskString }

var userDisallowTask = &UserDisallowTask{TaskString: TaskUserDisallow}

var _ tasks.Task = (*UserDisallowTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*UserDisallowTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*UserDisallowTask)(nil)

func (UserDisallowTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationBlogUserDisallowEmail")
}

func (UserDisallowTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationBlogUserDisallowEmail")
	return &v
}

func (UserDisallowTask) Action(w http.ResponseWriter, r *http.Request) any {
	UsersPermissionsDisallowPage(w, r)
	return nil
}

func (UserDisallowTask) TargetUserIDs(evt eventbus.TaskEvent) ([]int32, error) {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}, nil
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}, nil
	}
	return nil, fmt.Errorf("target user id not provided")
}

func (UserDisallowTask) TargetEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("deleteUserRoleEmail")
}

func (UserDisallowTask) TargetInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("delete_user_role")
	return &v
}
