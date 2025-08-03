package writings

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
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
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	permid, err := strconv.Atoi(r.PostFormValue("permid"))
	if err != nil {
		return fmt.Errorf("permid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	id, username, role, err2 := roleInfoByPermID(r.Context(), queries, int32(permid))
	if err := queries.AdminDeleteUserRole(r.Context(), int32(permid)); err != nil {
		return fmt.Errorf("delete user role fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err2 == nil {
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
	} else {
		log.Printf("lookup role: %v", err2)
	}
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
