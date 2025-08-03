package writings

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// UserAllowTask grants a user a permission.
type UserAllowTask struct{ tasks.TaskString }

var userAllowTask = &UserAllowTask{TaskString: TaskUserAllow}

var _ tasks.Task = (*UserAllowTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*UserAllowTask)(nil)

func (UserAllowTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	username := r.PostFormValue("username")
	role := r.PostFormValue("role")
	u, err := queries.SystemGetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
	if err != nil {
		return fmt.Errorf("get user by username fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if err := queries.SystemCreateUserRole(r.Context(), db.SystemCreateUserRoleParams{
		UsersIdusers: u.Idusers,
		Name:         role,
	}); err != nil {
		return fmt.Errorf("create user role fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["targetUserID"] = u.Idusers
			evt.Data["Username"] = u.Username.String
			evt.Data["Role"] = role
		}
	}
	return nil
}

func (UserAllowTask) TargetUserIDs(evt eventbus.TaskEvent) ([]int32, error) {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}, nil
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}, nil
	}
	return nil, fmt.Errorf("target user id not provided")
}

func (UserAllowTask) TargetEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("setUserRoleEmail")
}

func (UserAllowTask) TargetInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("set_user_role")
	return &v
}
