package writings

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// UserDisallowTask removes a user's permission.
type UserDisallowTask struct{ tasks.TaskString }

var userDisallowTask = &UserDisallowTask{TaskString: TaskUserDisallow}

var _ tasks.Task = (*UserDisallowTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*UserDisallowTask)(nil)

func (UserDisallowTask) Action(w http.ResponseWriter, r *http.Request) any {
	if r.URL.Path == "/admin/writings/users/roles" {
		queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
		permid := r.PostFormValue("permid")
		permidi, err := strconv.Atoi(permid)
		if err != nil {
			return fmt.Errorf("perm id parse: %w", err)
		}
		if err := queries.DeleteUserRole(r.Context(), int32(permidi)); err != nil {
			return fmt.Errorf("permissionUserDisallow: %w", err)
		}
		return nil
	}
	// TODO: inline UsersPermissionsDisallowPage or convert to a task page if
	// proxying is not required.
	return http.HandlerFunc(UsersPermissionsDisallowPage)
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
