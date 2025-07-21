package linker

import (
	"context"
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"
	"strings"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

func AdminUserRolesPage(w http.ResponseWriter, r *http.Request) {
	type PermissionUser struct {
		*db.GetUserRolesRow
		Username sql.NullString
		Email    sql.NullString
	}

	type Data struct {
		*common.CoreData
		UserLevels []*PermissionUser
		Search     string
		Roles      []*db.Role
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Search:   r.URL.Query().Get("search"),
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if roles, err := data.AllRoles(); err == nil {
		data.Roles = roles
	}
	rows, err := queries.GetUserRoles(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getUsersPermissions Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	var perms []*PermissionUser
	for _, p := range rows {
		row, err := queries.GetUserById(r.Context(), p.UsersIdusers)
		if err != nil {
			log.Printf("GetUserById Error: %s", err)
			continue
		}
		perms = append(perms, &PermissionUser{GetUserRolesRow: p, Username: row.Username, Email: row.Email})
	}

	if data.Search != "" {
		q := strings.ToLower(data.Search)
		var filtered []*PermissionUser
		for _, row := range perms {
			if strings.Contains(strings.ToLower(row.Username.String), q) {
				filtered = append(filtered, row)
			}
		}
		perms = filtered
	}
	data.UserLevels = perms

	handlers.TemplateHandler(w, r, "adminUserRolesPage.gohtml", data)
}

type userAllowTask struct{ tasks.TaskString }

var UserAllowTask = &userAllowTask{TaskString: TaskUserAllow}

var _ notif.TargetUsersNotificationProvider = (*userAllowTask)(nil)
var _ tasks.Task = (*userAllowTask)(nil)

func (userAllowTask) Action(w http.ResponseWriter, r *http.Request) {
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
		u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
		if err != nil {
			log.Printf("GetUserByUsername Error: %s", err)
			continue
		}
		if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{
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
	handlers.TaskDoneAutoRefreshPage(w, r)
}

type userDisallowTask struct{ tasks.TaskString }

var UserDisallowTask = &userDisallowTask{TaskString: TaskUserDisallow}

var _ notif.TargetUsersNotificationProvider = (*userDisallowTask)(nil)
var _ tasks.Task = (*userDisallowTask)(nil)

func (userDisallowTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	r.ParseForm()
	ids := r.Form["permids"]
	if len(ids) == 0 {
		if id := r.PostFormValue("permid"); id != "" {
			ids = append(ids, id)
		}
	}
	for _, idStr := range ids {
		permid, _ := strconv.Atoi(idStr)
		infoID, username, role, err2 := roleInfoByPermID(r.Context(), queries, int32(permid))
		if err := queries.DeleteUserRole(r.Context(), int32(permid)); err != nil {
			log.Printf("permissionUserDisallow Error: %s", err)
		} else if err2 == nil {
			if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
				if evt := cd.Event(); evt != nil {
					if evt.Data == nil {
						evt.Data = map[string]any{}
					}
					evt.Data["targetUserID"] = infoID
					evt.Data["Username"] = username
					evt.Data["Role"] = role
				}
			}
		}
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}

func roleInfoByPermID(ctx context.Context, q *db.Queries, id int32) (int32, string, string, error) {
	rows, err := q.GetPermissionsWithUsers(ctx, db.GetPermissionsWithUsersParams{Username: sql.NullString{}})
	if err != nil {
		return 0, "", "", err
	}
	for _, row := range rows {
		if row.IduserRoles == id {
			return row.UsersIdusers, row.Username.String, row.Name, nil
		}
	}
	return 0, "", "", sql.ErrNoRows
}

func (userAllowTask) TargetUserIDs(evt eventbus.TaskEvent) []int32 {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}
	}
	return nil
}

func (userAllowTask) TargetEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("setUserRoleEmail")
}

func (userAllowTask) TargetInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("set_user_role")
	return &v
}

func (userDisallowTask) TargetUserIDs(evt eventbus.TaskEvent) []int32 {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}
	}
	return nil
}

func (userDisallowTask) TargetEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("deleteUserRoleEmail")
}

func (userDisallowTask) TargetInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("delete_user_role")
	return &v
}
