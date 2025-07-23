package admin

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// NewsUserAllowTask grants a role to a user and notifies admins.
type NewsUserAllowTask struct{ tasks.TaskString }

// TaskNewsUserAllow identifies a request to grant a user a role.
const TaskNewsUserAllow tasks.TaskString = "allow"

var newsUserAllow = &NewsUserAllowTask{TaskString: TaskNewsUserAllow}

var _ tasks.Task = (*NewsUserAllowTask)(nil)
var _ tasks.AuditableTask = (*NewsUserAllowTask)(nil)
var _ notifications.AdminEmailTemplateProvider = (*NewsUserAllowTask)(nil)
var _ notifications.TargetUsersNotificationProvider = (*NewsUserAllowTask)(nil)

func (NewsUserAllowTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	username := r.PostFormValue("username")
	role := r.PostFormValue("role")
	u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
	if err != nil {
		return fmt.Errorf("get user by username fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{
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
