package news

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// UserAllowTask grants a user a role.
type UserAllowTask struct{ tasks.TaskString }

var userAllowTask = &UserAllowTask{TaskString: TaskUserAllow}

var _ tasks.Task = (*UserAllowTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*UserAllowTask)(nil)
var _ tasks.EmailTemplatesRequired = (*UserAllowTask)(nil)

func (UserAllowTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationNewsUserAllow.EmailTemplates(), true
}

func (UserAllowTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationNewsUserAllow.NotificationTemplate()
	return &v
}

func (UserAllowTask) RequiredTemplates() []tasks.Template {
	return append([]tasks.Template{tasks.Template(handlers.TemplateRunTaskPage)},
		EmailTemplateAdminNotificationNewsUserAllow.RequiredTemplates()...)
}

func (UserAllowTask) Action(w http.ResponseWriter, r *http.Request) any {
	username := r.PostFormValue("username")
	role := r.PostFormValue("role")
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: "/news",
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if err := cd.AllowNewsUser(username, role); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("permissionUserAllow: %w", err).Error())
	}

	return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
}
