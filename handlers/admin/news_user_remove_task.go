package admin

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// NewsUserRemoveTask revokes a role from a user and notifies admins.
type NewsUserRemoveTask struct{ tasks.TaskString }

// TaskNewsUserRemove identifies a request to revoke a user's role.
const TaskNewsUserRemove tasks.TaskString = "remove"

var newsUserRemove = &NewsUserRemoveTask{TaskString: TaskNewsUserRemove}

var _ tasks.Task = (*NewsUserRemoveTask)(nil)
var _ tasks.AuditableTask = (*NewsUserRemoveTask)(nil)
var _ notifications.AdminEmailTemplateProvider = (*NewsUserRemoveTask)(nil)
var _ notifications.TargetUsersNotificationProvider = (*NewsUserRemoveTask)(nil)

func (NewsUserRemoveTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	permid, err := strconv.Atoi(r.PostFormValue("permid"))
	if err != nil {
		return fmt.Errorf("permid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	id, username, role, err := roleInfoByPermID(r.Context(), queries, int32(permid))
	if err != nil {
		log.Printf("lookup role: %v", err)
	}
	if err := queries.DeleteUserRole(r.Context(), int32(permid)); err != nil {
		return fmt.Errorf("delete user role fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err == nil {
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["targetUserID"] = id
				evt.Data["Username"] = username
				evt.Data["Role"] = role
			}
		}
	}
	return nil
}
