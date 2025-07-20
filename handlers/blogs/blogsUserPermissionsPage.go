package blogs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"

	common "github.com/arran4/goa4web/core/common"
	db "github.com/arran4/goa4web/internal/db"

	"log"
	"net/http"
	"strconv"
	"strings"

	handlers "github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// UserAllowTask grants a user a permission.
type UserAllowTask struct{ tasks.TaskString }

var userAllowTask = &UserAllowTask{TaskString: TaskUserAllow}

var _ tasks.Task = (*UserAllowTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*UserAllowTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*UserAllowTask)(nil)

func (UserAllowTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationBlogUserAllowEmail")
}

func (UserAllowTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationBlogUserAllowEmail")
	return &v
}

func (UserAllowTask) Action(w http.ResponseWriter, r *http.Request) {
	UsersPermissionsPermissionUserAllowPage(w, r)
}

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

func (UserDisallowTask) Action(w http.ResponseWriter, r *http.Request) {
	UsersPermissionsDisallowPage(w, r)
}

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

func (UsersAllowTask) Action(w http.ResponseWriter, r *http.Request) {
	UsersPermissionsBulkAllowPage(w, r)
}

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

func (UsersDisallowTask) Action(w http.ResponseWriter, r *http.Request) {
	UsersPermissionsBulkDisallowPage(w, r)
}

func GetPermissionsByUserIdAndSectionBlogsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !(cd.HasRole("content writer") || cd.HasRole("administrator")) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	type Data struct {
		*common.CoreData
		Rows   []*db.GetUserRolesRow
		Filter string
		Roles  []*db.Role
	}

	data := Data{
		CoreData: cd,
		Filter:   r.URL.Query().Get("level"),
	}

	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	if roles, err := cd.AllRoles(); err == nil {
		data.Roles = roles
	}

	rows, err := queries.GetUserRoles(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	if data.Filter != "" {
		filtered := rows[:0]
		for _, row := range rows {
			if row.Role == data.Filter {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}

	data.Rows = rows

	handlers.TemplateHandler(w, r, "userPermissionsPage.gohtml", data)
}

func UsersPermissionsPermissionUserAllowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	username := r.PostFormValue("username")
	level := r.PostFormValue("role")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/blogs/bloggers",
	}
	if u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetUserByUsername: %w", err).Error())
	} else if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{
		UsersIdusers: u.Idusers,
		Name:         level,
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("permissionUserAllow: %w", err).Error())
	} else if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["targetUserID"] = u.Idusers
			evt.Data["Username"] = u.Username.String
			evt.Data["Role"] = level
		}
	}

	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func UsersPermissionsDisallowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	permid := r.PostFormValue("permid")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/blogs/bloggers",
	}
	if permidi, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else {
		id, username, role, err2 := roleInfoByPermID(r.Context(), queries, int32(permidi))
		if err := queries.DeleteUserRole(r.Context(), int32(permidi)); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("CreateLanguage: %w", err).Error())
		} else if err2 == nil {
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
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func UsersPermissionsBulkAllowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	names := strings.FieldsFunc(r.PostFormValue("usernames"), func(r rune) bool { return r == ',' || r == '\n' || r == ' ' || r == '\t' })
	level := r.PostFormValue("role")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/blogs/bloggers",
	}

	for _, n := range names {
		if n == "" {
			continue
		}
		u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: n})
		if err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("GetUserByUsername %s: %w", n, err).Error())
			continue
		}
		if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{
			UsersIdusers: u.Idusers,
			Name:         level,
		}); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("permissionUserAllow %s: %w", n, err).Error())
		}
	}

	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func UsersPermissionsBulkDisallowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	permids := r.PostForm["permid"]
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/blogs/bloggers",
	}

	for _, id := range permids {
		if id == "" {
			continue
		}
		permidi, err := strconv.Atoi(id)
		if err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi %s: %w", id, err).Error())
			continue
		}
		if err := queries.DeleteUserRole(r.Context(), int32(permidi)); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("permissionUserDisallow %s: %w", id, err).Error())
		}
	}

	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
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

func (UserAllowTask) TargetUserIDs(evt eventbus.Event) []int32 {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}
	}
	return nil
}

func (UserAllowTask) TargetEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("setUserRoleEmail")
}

func (UserAllowTask) TargetInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("set_user_role")
	return &v
}

func (UserDisallowTask) TargetUserIDs(evt eventbus.Event) []int32 {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}
	}
	return nil
}

func (UserDisallowTask) TargetEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("deleteUserRoleEmail")
}

func (UserDisallowTask) TargetInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("delete_user_role")
	return &v
}
