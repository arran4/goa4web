package languages

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// CreateLanguageTask creates a new language.
type CreateLanguageTask struct{ tasks.TaskString }

var createLanguageTask = &CreateLanguageTask{TaskString: tasks.TaskString("Create Language")}

var _ tasks.Task = (*CreateLanguageTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*CreateLanguageTask)(nil)

func (CreateLanguageTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cname := r.PostFormValue("cname")
	res, err := queries.InsertLanguage(r.Context(), sql.NullString{String: cname, Valid: true})
	if err != nil {
		return fmt.Errorf("create language fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if id, err := res.LastInsertId(); err == nil {
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["LanguageID"] = id
				evt.Data["LanguageName"] = cname
			}
		}
	}
	return nil
}

func (CreateLanguageTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationLanguageCreateEmail")
}

func (CreateLanguageTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationLanguageCreateEmail")
	return &v
}
