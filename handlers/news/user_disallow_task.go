package news

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// UserDisallowTask removes a user's role.
type UserDisallowTask struct{ tasks.TaskString }

var userDisallowTask = &UserDisallowTask{TaskString: TaskUserDisallow}

var _ tasks.Task = (*UserDisallowTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*UserDisallowTask)(nil)
var _ tasks.EmailTemplatesRequired = (*UserDisallowTask)(nil)

func (UserDisallowTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationNewsUserDisallow.EmailTemplates(), true
}

func (UserDisallowTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationNewsUserDisallow.NotificationTemplate()
	return &v
}

func (UserDisallowTask) RequiredTemplates() []tasks.Template {
	return append([]tasks.Template{tasks.Template(handlers.TemplateRunTaskPage)},
		EmailTemplateAdminNotificationNewsUserDisallow.RequiredTemplates()...)
}

func (UserDisallowTask) Action(w http.ResponseWriter, r *http.Request) any {
	permid := r.PostFormValue("permid")
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: "/news",
	}
	if permidi, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := r.Context().Value(consts.KeyCoreData).(*common.CoreData).DisallowNewsUser(int32(permidi)); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("CreateLanguage: %w", err).Error())
	}
	return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
}
