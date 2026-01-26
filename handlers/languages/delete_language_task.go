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

// DeleteLanguageTask removes a language entry.
type DeleteLanguageTask struct{ tasks.TaskString }

var deleteLanguageTask = &DeleteLanguageTask{TaskString: tasks.TaskString("Delete Language")}

var _ tasks.Task = (*DeleteLanguageTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*DeleteLanguageTask)(nil)
var _ tasks.EmailTemplatesRequired = (*DeleteLanguageTask)(nil)

func (DeleteLanguageTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasAdminRole() {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		})
	}
	id, name, err := cd.DeleteLanguage(r.PostFormValue("cid"))
	if err != nil {
		return fmt.Errorf("delete language fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["LanguageID"] = id
			evt.Data["LanguageName"] = name
		}
	}
	return nil
}

func (DeleteLanguageTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationLanguageDelete.EmailTemplates(), true
}

func (DeleteLanguageTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationLanguageDelete.NotificationTemplate()
	return &v
}

func (DeleteLanguageTask) EmailTemplatesRequired() []tasks.Page {
	return EmailTemplateAdminNotificationLanguageDelete.RequiredPages()
}
