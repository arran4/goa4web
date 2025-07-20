package blogs

import (
	"database/sql"
	"errors"
	"fmt"

	common "github.com/arran4/goa4web/core/common"
	db "github.com/arran4/goa4web/internal/db"

	"net/http"
	"strconv"
	"strings"

	handlers "github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// UserAllowTask grants a user a permission.
type UserAllowTask struct{ tasks.TaskString }

var userAllowTask = &UserAllowTask{TaskString: TaskUserAllow}

var _ tasks.Task = (*UserAllowTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*UserAllowTask)(nil)

// AdminEmailTemplate ensures admins receive a consistent message when
// permissions are granted so they stay informed about changes.
func (UserAllowTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("blogPermissionEmail")
}

// AdminInternalNotificationTemplate pairs with the admin email so the
// dashboard also reflects the permission update.
func (UserAllowTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("blog_permission")
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

// AdminEmailTemplate notifies administrators when a blog permission is
// revoked, keeping them in the loop if access is removed.
func (UserDisallowTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("blogPermissionEmail")
}

// AdminInternalNotificationTemplate mirrors the email so in-app alerts show
// the permission revocation too.
func (UserDisallowTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("blog_permission")
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

// AdminEmailTemplate informs admins when many users receive access so they
// can audit bulk changes easily.
func (UsersAllowTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("blogPermissionEmail")
}

// AdminInternalNotificationTemplate complements the email with an in-app
// notice for bulk permissions.
func (UsersAllowTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("blog_permission")
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

// AdminEmailTemplate alerts administrators when multiple users have
// permissions removed so there is an audit trail.
func (UsersDisallowTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("blogPermissionEmail")
}

// AdminInternalNotificationTemplate sends the same message inside the admin
// dashboard for visibility.
func (UsersDisallowTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("blog_permission")
	return &v
}

func (UsersDisallowTask) Action(w http.ResponseWriter, r *http.Request) {
	UsersPermissionsBulkDisallowPage(w, r)
}

func GetPermissionsByUserIdAndSectionBlogsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)
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

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
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
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	username := r.PostFormValue("username")
	level := r.PostFormValue("role")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
		Back:     "/blogs/bloggers",
	}
	if u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetUserByUsername: %w", err).Error())
	} else if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{
		UsersIdusers: u.Idusers,
		Name:         level,
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("permissionUserAllow: %w", err).Error())
	}

	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func UsersPermissionsDisallowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	permid := r.PostFormValue("permid")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
		Back:     "/blogs/bloggers",
	}
	if permidi, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.DeleteUserRole(r.Context(), int32(permidi)); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("CreateLanguage: %w", err).Error())
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func UsersPermissionsBulkAllowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	names := strings.FieldsFunc(r.PostFormValue("usernames"), func(r rune) bool { return r == ',' || r == '\n' || r == ' ' || r == '\t' })
	level := r.PostFormValue("role")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
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
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	permids := r.PostForm["permid"]
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
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
