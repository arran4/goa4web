package linker

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// userAllowTask grants a user a role.
type userAllowTask struct{ tasks.TaskString }

// UserAllowTask is the exported instance used when registering routes.
var UserAllowTask = &userAllowTask{TaskString: TaskUserAllow}

var _ tasks.Task = (*userAllowTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*userAllowTask)(nil)

func (userAllowTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	usernames := r.PostFormValue("usernames")
	role := r.PostFormValue("role")
	fields := strings.FieldsFunc(usernames, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	for _, username := range fields {
		if username == "" {
			continue
		}
		u, err := queries.SystemGetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
		if err != nil {
			log.Printf("SystemGetUserByUsername Error: %s", err)
			continue
		}
		if err := queries.SystemCreateUserRole(r.Context(), db.SystemCreateUserRoleParams{
			UsersIdusers: u.Idusers,
			Name:         role,
		}); err != nil {
			log.Printf("permissionUserAllow Error: %s", err)
		} else if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["targetUserID"] = u.Idusers
				evt.Data["Username"] = u.Username.String
				evt.Data["Role"] = role
			}
		}
	}
	return nil
}

func (userAllowTask) TargetUserIDs(evt eventbus.TaskEvent) ([]int32, error) {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}, nil
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}, nil
	}
	return nil, fmt.Errorf("target user id not provided")
}

func (userAllowTask) TargetEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("setUserRoleEmail")
}

func (userAllowTask) TargetInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("set_user_role")
	return &v
}
