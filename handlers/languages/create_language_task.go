package languages

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

// CreateLanguageTask creates a new language.
type CreateLanguageTask struct{ tasks.TaskString }

const (
	EmailTemplateAdminNotificationLanguageCreate notif.EmailTemplateName = "adminNotificationLanguageCreateEmail"
)

var createLanguageTask = &CreateLanguageTask{TaskString: tasks.TaskString("Create Language")}

var _ tasks.Task = (*CreateLanguageTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*CreateLanguageTask)(nil)
var _ tasks.EmailTemplatesRequired = (*CreateLanguageTask)(nil)

func (CreateLanguageTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasAdminRole() {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		})
	}
	cname := r.PostFormValue("cname")
	id, err := cd.CreateLanguage("", cname)
	if err != nil {
		return fmt.Errorf("create language fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["LanguageID"] = id
		evt.Data["LanguageName"] = cname
	}
	return nil
}

func (CreateLanguageTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationLanguageCreate.EmailTemplates(), true
}

func (CreateLanguageTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationLanguageCreate.NotificationTemplate()
	return &v
}

func (CreateLanguageTask) EmailTemplatesRequired() []tasks.Page {
	return EmailTemplateAdminNotificationLanguageCreate.RequiredPages()
}
