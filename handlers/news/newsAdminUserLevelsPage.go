package news

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

func NewsAdminUserLevelsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		UserLevels []*db.GetUserRolesRow
		Roles      []*db.Role
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}

	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
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
	data.UserLevels = rows

	handlers.TemplateHandler(w, r, "adminUserLevelsPage.gohtml", data)
}

type newsUserAllowTask struct{ tasks.TaskString }

var _ tasks.Task = (*newsUserAllowTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*newsUserAllowTask)(nil)

func (newsUserAllowTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationNewsUserAllowEmail")
}

func (newsUserAllowTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsUserAllowEmail")
	return &v
}

func (newsUserAllowTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	username := r.PostFormValue("username")
	level := r.PostFormValue("role")
	u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
	if err != nil {
		log.Printf("GetUserByUsername Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{
		UsersIdusers: u.Idusers,
		Name:         level,
	}); err != nil {
		log.Printf("permissionUserAllow Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}

type newsUserRemoveTask struct{ tasks.TaskString }

var _ tasks.Task = (*newsUserRemoveTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*newsUserRemoveTask)(nil)

func (newsUserRemoveTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationNewsUserDisallowEmail")
}

func (newsUserRemoveTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsUserDisallowEmail")
	return &v
}

func (newsUserRemoveTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	permid, err := strconv.Atoi(r.PostFormValue("permid"))
	if err != nil {
		log.Printf("strconv.Atoi(permid) Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := queries.DeleteUserRole(r.Context(), int32(permid)); err != nil {
		log.Printf("permissionUserDisallow Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}

var NewsUserAllowTask = newsUserAllowTask{TaskString: tasks.TaskString("allow")}
var NewsUserRemoveTask = newsUserRemoveTask{TaskString: tasks.TaskString("remove")}
