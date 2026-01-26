package languages

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

// RenameLanguageTask performs a language rename action.
type RenameLanguageTask struct{ tasks.TaskString }

var renameLanguageTask = &RenameLanguageTask{TaskString: tasks.TaskString("Rename Language")}

var _ tasks.Task = (*RenameLanguageTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*RenameLanguageTask)(nil)
var _ tasks.EmailTemplatesRequired = (*RenameLanguageTask)(nil)

func (RenameLanguageTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasAdminRole() {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		})
	}
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		return fmt.Errorf("cid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	cname := r.PostFormValue("cname")
	var oldName string
	if rows, err := cd.Languages(); err == nil {
		for _, l := range rows {
			if l.ID == int32(cid) {
				oldName = l.Nameof.String
				break
			}
		}
	}
	if err := cd.RenameLanguage(oldName, cname); err != nil {
		return fmt.Errorf("rename language fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["LanguageID"] = cid
			evt.Data["LanguageName"] = cname
		}
	}
	return nil
}

func (RenameLanguageTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationLanguageRename.EmailTemplates(), true
}

func (RenameLanguageTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationLanguageRename.NotificationTemplate()
	return &v
}

func (RenameLanguageTask) EmailTemplatesRequired() []tasks.Page {
	return EmailTemplateAdminNotificationLanguageRename.RequiredPages()
}
